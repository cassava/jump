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

// The reason for this is that stdout is evaluated by the shell, and we only
// want one thing evaluated. Everything else should go to stderr.
// I don't know what other packages here might somehow print to stdout;
// I'm trying my best to prevent it. May or may not work.
var stdout = os.Stdout

func init() {
	os.Stdout = os.Stderr
}

var (
	source bool
	create bool
	modify bool
	remove bool
)

func init() {
	flag.BoolVar(&source, "source", false, "create shell function with name")
	flag.BoolVarP(&create, "create", "c", false, "create a new jump point")
	flag.BoolVarP(&modify, "modify", "m", false, "modify a jump point")
	flag.BoolVarP(&remove, "remove", "r", false, "remove a jump point")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-cmr] profile [path]\n\n", os.Args[0])
		flag.PrintDefaults()
	}
}

var (
	ErrExists    = errors.New("named profile already exists")
	ErrNotExists = errors.New("named profile does not exist")
	ErrNotAbs    = errors.New("jump path should be absolute")
)

func profile(name string) string {
	return "jump/" + name
}

func printFn(name string) error {
	fmt.Fprintf(stdout, "function %s() { eval $(%s $@) }\n", name, os.Args[0])
	return nil
}

func listJP() error {
	type entry struct{ Name, Path string }
	all := []entry{}

	err := xdg.MergeConfigR("jump", func(dir string) error {
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
		fmt.Fprintf(os.Stderr, format, e.Name, e.Path)
	}
	return nil
}

func createJP(name, dst string) error {
	// Make sure the configuration directory exists
	dirpath := xdg.UserConfig("jump")
	if ex, _ := osutil.DirExists(dirpath); !ex {
		os.MkdirAll(dirpath, 0777)
	}

	// Make sure the profile does not exist
	jp := xdg.UserConfig(profile(name))
	if ex, _ := osutil.FileExists(jp); ex {
		return ErrExists
	}

	// Create a jump point
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
	fmt.Fprintf(stdout, "cd %q\n", dst)
	return nil
}

func main() {
	flag.Parse()

	// Make sure at most one option is specified
	toggled := 0
	for _, opt := range []bool{source, create, modify, remove} {
		if opt {
			toggled++
		}
	}
	if toggled > 1 {
		usageError("multiple commands specified")
	}

	n, args := flag.NArg(), flag.Args()
	var err error
	if source {
		name := "jp"
		if n == 1 {
			name = args[0]
		} else if n > 1 {
			usageError("too many arguments")
		}
		err = printFn(name)
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
	} else if n == 0 {
		listJP()
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
