// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !plan9
// +build !plan9

/*
The getgo command is deprecated.

Deprecated: See https://go.dev/issues/60951.

Follow the instructions at https://go.dev/doc/install to install Go instead.

Tips:

To find the latest available go version, run:

	go list -m -f '{{.Version}}' go@latest

If you want to use the latest go by default, you can use "go env -w" to override "GOTOOLCHAIN":

	go env -w GOTOOLCHAIN=go$(go list -m -f '{{.Version}}' go@latest)+auto

See https://go.dev/blog/toolchain for more information about toolchain management.
*/
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	interactive = flag.Bool("i", false, "Interactive mode, prompt for inputs.")
	verbose     = flag.Bool("v", false, "Verbose.")
	setupOnly   = flag.Bool("skip-dl", false, "Don't download - only set up environment variables")
	goVersion   = flag.String("version", "", `Version of Go to install (e.g. "1.8.3"). If empty, uses the latest version.`)

	version = "devel"
)

var errExitCleanly error = errors.New("exit cleanly sentinel value")

func main() {
	flag.Parse()
	fmt.Fprintln(os.Stderr, "getgo is deprecated. See https://pkg.go.dev/golang.org/x/tools/cmd/getgo.")

	if *goVersion != "" && !strings.HasPrefix(*goVersion, "go") {
		*goVersion = "go" + *goVersion
	}

	ctx := context.Background()

	verbosef("version " + version)

	runStep := func(s step) {
		err := s(ctx)
		if err == errExitCleanly {
			os.Exit(0)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}

	if !*setupOnly {
		runStep(welcome)
		runStep(checkOthers)
		runStep(chooseVersion)
		runStep(downloadGo)
	}

	runStep(setupGOPATH)
}

func verbosef(format string, v ...interface{}) {
	if !*verbose {
		return
	}

	fmt.Printf(format+"\n", v...)
}

func prompt(ctx context.Context, query, defaultAnswer string) (string, error) {
	if !*interactive {
		return defaultAnswer, nil
	}

	fmt.Printf("%s [%s]: ", query, defaultAnswer)

	type result struct {
		answer string
		err    error
	}
	ch := make(chan result, 1)
	go func() {
		s := bufio.NewScanner(os.Stdin)
		if !s.Scan() {
			ch <- result{"", s.Err()}
			return
		}
		answer := s.Text()
		if answer == "" {
			answer = defaultAnswer
		}
		ch <- result{answer, nil}
	}()

	select {
	case r := <-ch:
		return r.answer, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func runCommand(ctx context.Context, prog string, args ...string) ([]byte, error) {
	verbosef("Running command: %s %v", prog, args)

	cmd := exec.CommandContext(ctx, prog, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("running cmd '%s %s' failed: %s err: %v", prog, strings.Join(args, " "), string(out), err)
	}
	if out != nil && err == nil && len(out) != 0 {
		verbosef("%s", out)
	}

	return out, nil
}
