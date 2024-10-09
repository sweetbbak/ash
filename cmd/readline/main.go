package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/alecthomas/chroma/quick"
	"github.com/kballard/go-shellquote"
	"github.com/reeflective/readline"
	"github.com/reeflective/readline/inputrc"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

type ash struct {
	highlight func(line []rune) string
	multiLine func(line []rune) (accept bool)
	shell     *readline.Shell
	exes      []string
}

func acceptMultiline(line []rune) (accept bool) {
	// Errors are either: unterminated quotes, or unterminated escapes.
	_, err := shellquote.Split(string(line))
	if err == nil {
		return true
	}

	switch err {
	case shellquote.UnterminatedDoubleQuoteError, shellquote.UnterminatedSingleQuoteError:
		// Unterminated quotes: keep reading.
		return false

	case shellquote.UnterminatedEscapeError:
		// If there no are trailing spaces, keep reading.
		if len(line) > 0 && line[len(line)-1] == '\\' {
			return false
		}
		return true
	}

	return true
}

// highlighter func for line highlighting
func (a *ash) highlighter(line []rune) string {
	var highlighter strings.Builder

	err := quick.Highlight(&highlighter, string(line), "bash", "terminal16m", "witchhazel")
	if err != nil {
		return string(line)
	}

	// hiline := highlighter.String()
	// words := strings.Split(hiline, " ")
	words := strings.Split(string(line), " ")

	for i, word := range words {
		_, err := os.Stat(word)
		if err == nil {
			words[i] = fmt.Sprintf("%s%s%s", UL, word, CLEAR)
		}

		for _, e := range a.exes {
			e = strings.TrimSpace(e)
			word = strings.TrimSpace(word)
			if e == word {
				words[i] = fmt.Sprintf("\x1b[0m%s%s%s", GREEN, word, CLEAR)
			}
		}
	}

	return strings.Join(words, " ")
}

var (
	GREEN string = "\x1b[32m"
	CLEAR string = "\x1b[0m"
	UL    string = "\x1b[4m"
)

var _, EXES = lookPath()

func (a *ash) Prompt() string {
	cwd, _ := os.Getwd()
	cwd = path.Base(cwd)
	host, _ := os.Hostname()
	usr, _ := user.Current()

	return fmt.Sprintf("\x1b[32m%s@%s \x1b[34m%s $\x1b[0m ", usr.Username, host, cwd)
}

func (a *ash) Run() error {
	// shell := readline.NewShell(inputrc.WithName("ash"))
	config := a.shell.Config

	// Or adding it later and re-reading the inputrc file (not recommended)
	// shell.Opts = append(shell.Options, inputrc.WithName(appname))
	a.shell.Config.ReadFile("inputrc")
	// Access the map of all options
	// allOptions := config.Vars

	// Querying option values
	// editingMode := config.GetString("editing-mode")        // Either "vi" or "emacs"
	// autocomplete := config.GetBool("autocomplete")         // directly as boolean (on/off)
	// viPromptMode := config.GetString("vi-ins-mode-string") // Some options are string values
	// keyTimeout := config.GetInt("keyseq-timeout")          // Others are integers

	// Modifying the option values
	config.Set("editing-mode", "vi")
	config.Set("autocomplete", true)
	config.Set("keyseq-timeout", 100)

	// Or create and bind it in one call.
	a.shell.History.AddFromFile("history name", "ash_history")
	a.shell.SyntaxHighlighter = a.highlighter
	a.shell.AcceptMultiline = a.multiLine

	// a.shell.Prompt.Primary(func() string { return "$ " })
	a.shell.Prompt.Primary(a.Prompt)
	a.shell.Prompt.Right(func() string {
		now := time.Now()
		h := now.Hour()
		m := now.Minute()

		return fmt.Sprintf("%v:%v", h, m)
	})

	r, err := interp.New(interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
	if err != nil {
		log.Fatal(err)
	}

	parser := syntax.NewParser(syntax.Variant(syntax.LangBash))

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-signals:
			cancel()
			return
		case <-ctx.Done():
			return
		}
	}()

	// fmt.Fprintf(os.Stdout, "$ ")
	// parser.Interactive(os.Stdin, func(stmts []*syntax.Stmt) bool {
	// 	if parser.Incomplete() {
	// 		fmt.Fprintf(os.Stdout, "> ")
	// 		return true
	// 	}
	//
	// 	ctx := context.Background()
	// 	for _, stmt := range stmts {
	// 		err = r.Run(ctx, stmt)
	// 		if r.Exited() {
	// 			return false
	// 		}
	// 	}
	//
	// 	fmt.Fprintf(os.Stdout, "$ ")
	// 	return true
	// },
	// )

	for {
		line, err := a.shell.Readline()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// sfile, _ := parser.Parse(strings.NewReader(line), "ash")
		// r.Run(context.Background(), sfile)

		// var words []*syntax.Word
		// parser.Words(strings.NewReader(line), func(w *syntax.Word) bool {
		// 	words = append(words, w)
		// 	return true
		// })

		if err := parser.Stmts(strings.NewReader(line), func(stmt *syntax.Stmt) bool {
			if parser.Incomplete() {
				return true
			}

			r.Run(ctx, stmt)
			if r.Exited() {
				return false
			}

			return false
		}); err != nil {
			log.Println(err)
		}
	}

	return nil
}

func main() {
	_, exes := lookPath()
	a := ash{
		multiLine: acceptMultiline,
		// highlight: highlighter,
		shell: readline.NewShell(inputrc.WithName("ash")),
		exes:  exes,
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
