# `fileurl`

The `fileurl` package was inspired by the confusion seen in [`kubernetes/minikube#1310`](https://github.com/kubernetes/minikube/issues/1310). After referencing that issue, and the three links mentioned by [`@kichristensen`](https://github.com/kubernetes/minikube/issues/1310#issuecomment-325157114), my opinion is that the core problem is it's harder than people realize to transform between local file paths and URLs in the file scheme. The `fileurl` package addresses this by introducing of `fileurl.ToLocal` and `fileurl.FromLocal`. In addition, `fileurl.ToLocalSloppy` maps an invalid URL containing a drive-like Host to a reasonable path in a manner that would make the current `minikube` code work.

While `fileurl` contains working code, this repository was created solely to better document the problems here. I have not vetted it for further use, and have little intention of maintaining this long term. Feel free to learn from, fork, or even just copy the code. While I have applied the MIT license to this repository, there's little enough code that I'm not convinced it's protected by copyright.

## Details of the problem

Elsewhere go has been blamed for mishandling URLs with a path that looks like a windows path. The go maintainers examined the case of a correctly formed URL on [`golang/go#6027`](https://github.com/golang/go/issues/6027) and decided to leave this behavior as is. It reportedly agrees with Python, Node.JS, and Java, only disagreeing with Mono. Mono appears to be special casing this, going beyond [RFC 3986](https://tools.ietf.org/html/rfc3986)'s requirements.

The four ways the URL could appear, as I understand them, are as follows. Below each example, I show [how `net/url` parses them](https://play.golang.org/p/r11LJokbtdY).

|    | Two slashes | Three Slashes |
|---:|-------------|---------------|
|**Forward slashes**| `file://c:/windows/notepad.exe` | `file:///c:/windows/notepad.exe` |
| *Host* | `c:` | *empty* |
| *Path* | `/windows/notepad.exe` | `/c:/windows/notepad.exe` |
|**Reverse slashes**| `file://c:\windows\notepad.exe` | `file:///c:\windows\notepad.exe` |
| *Host* | *error* | *empty* |
| *Path* | *error* | `/c:\windows\notepad.exe` |

## How this came up

In the wild, code of the following form attempted to create a URL string. It happens to work on paths that start with a slash, such as those on Linux-like platforms, but creates invalid URLs from absolute Windows paths. (Both [Wikipedia](https://en.wikipedia.org/wiki/File_URI_scheme) and [a Microsoft blog post](https://blogs.msdn.microsoft.com/ie/2006/12/06/file-uris-in-windows/) clearly describe a two-slash prefix as invalid for local files.) It was then consumed by parsing the string as a URL and using the resulting URL's Path:

    return "file://" + filepath.ToSlash(somePath)
    ...
    u, _ := url.Parse(pathurl); os.Open(u.Path)

Another incorrect way to form a URL would be to leverage `net/url`'s URL type, and try to write code like the following:

    return url.URL{Scheme: "file", Path: filepath.ToSlash(somePath)}.String()

Unfortunately, this creates the same invalid URL string for a windows path. It has the slight advantage that if you pass the URL itself around, and refer only to its Path member, you may get the desired results. Note that if you omit the `filepath.ToSlash` call, you will see `%5C` substrings in place of the backslashes. Note as well that this use is problematic for relative paths in a Linux-like path; a path like `subdir/file` yields a URL string of `file://subdir/file` which would be parsed as referring to a host named subdir.

## What `fileurl` does

When forming a URL from a path, `fileurl` eschews relative paths. They're too hard to get right. Instead it converts to absolute paths. Then, if the path doesn't already start with a slash, it adds one.

When extracting a path from a URL, `fileurl` attempts to identify windows-like paths. As mentioned on [`golang/go#6027`](https://github.com/golang/go/issues/6027), this has risk of false positives; it's totally legal, but rare, to have a path like `/c:/foo` on a Linux-like OS. By using `fileurl`, you opt in to accepting that risk.

## Proposal for `minikube` and `docker/machine`

* `docker/machine` should alter [b2dReleaseGetter.download](https://github.com/docker/machine/blob/61ef47dc5d6b1658e3d6636f9382d50507c8c7e1/libmachine/mcnutils/b2d.go#L198-L199) to use logic equivalent to `fileurl.ToLocalSloppy`, optionally replacing with `fileurl.ToLocal` after clients have a chance to transition to using valid URLs.

* `kubernetes/minikube` should alter [DefaultDownloader.GetISOFileURI](https://github.com/kubernetes/minikube/blob/3cc68bd18b0726f303d6eb4786b8a4a997ceaa69/pkg/util/downloader.go#L51) to use logic equivalent to `fileurl.FromLocal`, once people have a chance to transition to the updated `docker/machine` URL support.

Unfortunately, there is an order dependency here, as valid URLs will be supported even more poorly than invalid ones by the current `docker/machine` code.
