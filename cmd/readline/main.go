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
	"path/filepath"
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
	highlight   func(line []rune) string
	multiLine   func(line []rune) (accept bool)
	shell       *readline.Shell
	parser      *syntax.Parser
	runner      *interp.Runner
	Stmts       *syntax.Stmt
	Executables []string
	PathDirs    []string
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

	hiline := highlighter.String()
	words := strings.Split(hiline, " ")
	// words := strings.Split(string(line), " ")

	for i, word := range words {
		word = strings.TrimSpace(word)
		words[i] = fmt.Sprintf("%s%s%s", UL, word, CLEAR)

		matches, _ := filepath.Glob(word + "*")
		for _, m := range matches {
			_, err := os.Stat(m)
			if err == nil {
				words[i] = fmt.Sprintf("%s%s%s", UL, word, CLEAR)
				break
			}
		}

		// _, err := os.Stat(word)
		// if err == nil {
		// 	words[i] = fmt.Sprintf("%s%s%s", UL, word, CLEAR)
		// }

		for _, e := range a.Executables {
			e = strings.TrimSpace(e)
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
	a.shell.History.AddFromFile("history name", ".ash_history")
	a.shell.SyntaxHighlighter = a.highlighter
	a.shell.AcceptMultiline = a.multiLine

	a.shell.Prompt.Primary(a.Prompt)
	a.shell.Prompt.Right(func() string {
		now := time.Now()
		h := now.Hour()
		m := now.Minute()

		return fmt.Sprintf("%v:%v", h, m)
	})

	runner, err := interp.New(interp.StdIO(os.Stdin, os.Stdout, os.Stderr), interp.Interactive(true))
	if err != nil {
		log.Fatal(err)
	}

	parser := syntax.NewParser(syntax.Variant(syntax.LangBash))

	a.parser = parser
	a.runner = runner

	// example of sourcing shell init files
	f, err := os.Open("shinit")
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	init, err := a.parser.Parse(f, "init")

	// signals := make(chan os.Signal, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	ctx, _ := context.WithCancel(context.Background())
	go func() {
		select {
		case <-signals:
			// cancel()
			return
		case <-ctx.Done():
			return
		}
	}()

	a.runner.Run(ctx, init)

	for {
		line, err := a.shell.Readline()
		if err == io.EOF {
			// break
		} else if err != nil {
			// return err
		}

		log.Println("parsing:", line)

		var (
			exitErr    error
			shouldExit bool
		)

		if err := a.parser.Stmts(strings.NewReader(line), func(stmt *syntax.Stmt) bool {
			exitErr = a.runner.Run(ctx, stmt)
			if a.runner.Exited() {
				shouldExit = true
			}

			return false
		}); err != nil {
			log.Println(err)
		}

		if e, ok := interp.IsExitStatus(exitErr); ok && shouldExit {
			os.Exit(int(e))
		}
	}

	return nil
}

func main() {
	logfile, _ := os.Create("ash.log")
	defer logfile.Close()

	log.SetOutput(logfile)

	_, exes := lookPath()
	a := ash{
		multiLine:   acceptMultiline,
		shell:       readline.NewShell(inputrc.WithName("ash")),
		Executables: exes,
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
