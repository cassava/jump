// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Command jp changes directory into one predefined.
//
// It can then be used in the following way:
//
//  jp dromi
//
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goulash/osutil"
	"github.com/goulash/xdg"
	flag "github.com/ogier/pflag"
)

var (
	create bool
	modify bool
	remove bool
)

func init() {
	flag.BoolVarP(&create, "create", "c", false, "create a new jump point")
	flag.BoolVarP(&modify, "modify", "m", false, "modify a jump point")
	flag.BoolVarP(&remove, "remove", "r", false, "remove a jump point")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-cmr] profile [path]\n", os.Args[0])
		flag.PrintDefaults()
	}
}

var (
	ErrExists    = errors.New("named profile already exists")
	ErrNotExists = errors.New("named profile does not exist")
	ErrNotAbs    = errors.New("jump path should be absolute")
)

var xdgName = "jp/"

func profile(name string) string {
	return "jp/" + name
}

func listJP() error {
	type entry struct{ Name, Path string }
	all := []entry{}

	err := xdg.MergeConfigR("jp", func(dir string) error {
		return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return nil
			}

			if info.Mode()&os.ModeSymlink == 0 {
				// Skip non-symlink files
				return nil
			}

			dst, err := os.Readlink(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return nil
			}
			all = append(all, entry{filepath.Base(path), dst})
			return nil
		})
	})
	if err != nil {
		return err
	}

	max := 0
	for _, e := range all {
		n := strings.Count(e.Name, "") - 1
		if max < n {
			max = n
		}
	}
	format := fmt.Sprintf("%%-%ds\t%%s\n", max)
	for _, e := range all {
		fmt.Printf(format, e.Name, e.Path)
	}
	return nil
}

func createJP(name, dst string) error {
	jp := xdg.UserConfig(profile(name))
	if ex, _ := osutil.FileExists(jp); ex {
		return ErrExists
	}
	dst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}
	return os.Symlink(dst, jp)
}

func modifyJP(name, dst string) error {
	jp := xdg.UserConfig(profile(name))
	_ = os.Remove(jp)
	return createJP(name, dst)
}

func removeJP(name string) error {
	jp := xdg.FindConfig(profile(name))
	if jp == "" {
		return ErrNotExists
	}
	return os.Remove(jp)
}

func jump(name string) error {
	jp := xdg.FindConfig(profile(name))
	if jp == "" {
		return ErrNotExists
	}
	dst, err := os.Readlink(jp)
	if err != nil {
		return err
	}
	return os.Chdir(dst)
}

func main() {
	flag.Parse()

	// Make sure at most one option is specified
	toggled := 0
	for _, opt := range []bool{create, modify, remove} {
		if opt {
			toggled++
		}
	}
	if toggled > 1 {
		usageError("multiple commands specified")
	}

	n, args := flag.NArg(), flag.Args()
	var err error
	if n == 0 {
		listJP()
	} else if create {
		if n != 2 {
			usageError("command create requires 2 arguments")
		}
		err = createJP(args[0], args[1])
	} else if modify {
		if n != 2 {
			usageError("command create requires 2 arguments")
		}
		err = modifyJP(args[0], args[1])
	} else if remove {
		err = removeJP(args[0])
	} else {
		err = jump(args[0])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func usageError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n\n", msg)
	flag.Usage()
	os.Exit(1)
}
