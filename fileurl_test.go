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
package fileurl_test

import (
	"net/url"
	"testing"

	"github.com/MichaelUrman/fileurl"
)

func TestFromLocal(t *testing.T) {
	var tests = []struct {
		Name  string
		Path  string
		URL   string
		Error error
	}{
		{"good", "c:/windows/notepad.exe", "file:///c:/windows/notepad.exe", nil},
		{"bad", "3:/windows/notepad.exe", "", fileurl.ErrRelative},
		{"weird", "/3:/windows/notepad.exe", "file:///3:/windows/notepad.exe", nil},
		{"remote", "//server/share/file", "", fileurl.ErrRemote},
		{"unix", "/usr/bin/vi", "file:///usr/bin/vi", nil},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			u, err := fileurl.FromLocal(tc.Path)
			if err != tc.Error {
				t.Errorf("error mismatch\n\thave: %v\n\twant: %v", err, tc.Error)
				return
			}

			if err == nil && (u == nil || u.String() != tc.URL) {
				t.Errorf("URL mismatch\n\thave: %q\n\twant: %q", u, tc.URL)
				return
			}
		})
	}
}

func TestToLocal(t *testing.T) {

	var tests = []struct {
		Name      string
		URL, Path string
		Error     error
		Sloppy    bool
	}{
		{"good", "file:///c:/windows/notepad.exe", "c:/windows/notepad.exe", nil, false},
		{"bad/sloppy", "file://c:/windows/notepad.exe", "c:/windows/notepad.exe", nil, true},
		{"bad/strict", "file://c:/windows/notepad.exe", "", fileurl.ErrRemote, false},
		{"weird", "file:///3:/windows/notepad.exe", "/3:/windows/notepad.exe", nil, false},
		{"weird/sloppy", "file://3:/windows/notepad.exe", "", fileurl.ErrRemote, true},
		{"remote", "file://filesrvr/share/windows/notepad.exe", "", fileurl.ErrRemote, false},
		// scheme/strict and scheme/sloppy return different errors due to implementation details; not contractual
		{"scheme/strict", "smb://filesrvr/share/windows/notepad.exe", "", fileurl.ErrRemote, false},
		{"scheme/sloppy", "smb://filesrvr/share/windows/notepad.exe", "", fileurl.ErrUnsupported, true},
		{"unix", "file:///usr/bin/vi", "/usr/bin/vi", nil, false},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			u, err := url.Parse(tc.URL)
			if err != nil {
				t.Errorf("unexpected error parsing url %q", tc.URL)
				return
			}

			to := fileurl.ToLocal
			if tc.Sloppy {
				to = fileurl.ToLocalSloppy
			}

			p, err := to(u)
			if err != tc.Error {
				t.Errorf("error mismatch\n\thave: %v\n\twant: %v", err, tc.Error)
				return
			}

			if p != tc.Path {
				t.Errorf("path mismatch\n\thave: %q\n\twant: %q", p, tc.Path)
				return
			}
		})
	}
}
