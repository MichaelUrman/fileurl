// Package fileurl makes it easier to round trip local file paths through a net/url.URL
//
// Copyright (c) 2019
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package fileurl

import (
	"errors"
	"net/url"
)

var (
	// ErrRelative is returned for relative paths or URLs
	ErrRelative = errors.New("path or URL is not absolute")

	// ErrRemote is returned for non-local path URLs
	ErrRemote = errors.New("path or url references remote location")

	// ErrUnsupported is returned for URLs with unsupported data
	ErrUnsupported = errors.New("url uses unsupported scheme, query, or fragment")
)

// FromLocal creates a file:///... url for a local file.
func FromLocal(path string) (*url.URL, error) {
	if path[:2] == "//" {
		return nil, ErrRemote
	}
	if path[0] != '/' {
		if len(path) >= 2 && path[1] == ':' && ((path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z')) {
			path = "/" + path
		} else {
			return nil, ErrRelative
		}
	}
	return url.Parse("file://" + path)
}

// ToLocal extracts a local path from a URL in the file scheme.
//
// If the URL appears to be to a remote resource, ToLocal wil return an error.
func ToLocal(u *url.URL) (string, error) {
	if u.Host != "" {
		return "", ErrRemote
	}
	return ToLocalSloppy(u)
}

// ToLocalSloppy extracts a local path from a URL in the file scheme. Unlike ToLocal, ToLocalSloppy
// will attempt to convert an invalid windows-style file://<drive>:/path URL to a local path.
//
// If the URL appears to be to a remote resource, ToLocalSloppy wil return an error.
func ToLocalSloppy(u *url.URL) (string, error) {
	if u.Scheme != "file" || u.RawQuery != "" || u.Fragment != "" || u.User != nil || u.Path == "" {
		return "", ErrUnsupported
	}

	p := u.Path

	// turn host=c: path=/path  into path=/c:/path
	if len(u.Host) != 0 {
		if len(u.Host) != 2 || u.Host[1] != ':' || !isDriveLetter(u.Host[0]) {
			return "", ErrRemote
		}
		p = "/" + u.Host + p
	}

	// turn /x:... into x:... for ascii letters in place of x.
	if len(p) >= 3 && p[0] == '/' && p[2] == ':' && isDriveLetter(p[1]) {
		p = p[1:]
	}

	return p, nil
}

func isDriveLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}
