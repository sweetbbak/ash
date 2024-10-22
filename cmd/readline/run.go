package main

import (
	"context"
	"io"
	"os"
	// "golang.org/x/term"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// func runAll() error {
// 	r, err := interp.New(interp.Interactive(true), interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
// 	if err != nil {
// 		return err
// 	}
//
// 	if *command != "" {
// 		return run(r, strings.NewReader(*command), "")
// 	}
//
// 	if flag.NArg() == 0 {
// 		if term.IsTerminal(int(os.Stdin.Fd())) {
// 			return runInteractive(r, os.Stdin, os.Stdout, os.Stderr)
// 		}
// 		return run(r, os.Stdin, "")
// 	}
// 	for _, path := range flag.Args() {
// 		if err := runPath(r, path); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func run(r *interp.Runner, reader io.Reader, name string) error {
	prog, err := syntax.NewParser().Parse(reader, name)
	if err != nil {
		return err
	}
	r.Reset()
	ctx := context.Background()
	return r.Run(ctx, prog)
}

func runPath(r *interp.Runner, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return run(r, f, path)
}
