package main

import (
	"log"

	"github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
	"gopkg.in/readline.v1"
)

const (
	PS1 = "\033[1;31m> \033[0m"
	PS2 = "\033[1;31m>> \033[0m"
)

// REPL is
type REPL struct {
	L       *lua.LState // the lua virtual machine
	toolbox *ToolBox
	reader  *readline.Instance
}

// NewREPL is create a instance of REPL
func NewREPL() *REPL {
	r := new(REPL)
	r.L = lua.NewState()
	r.toolbox = NewToolBox("/data")
	if reader, err := readline.New(PS1); err != nil {
		log.Println(err)
		return nil
	}
	r.reader = reader
	return r
}

// Close the REPL
func (p *REPL) Close() {
	p.toolbox.Close()
	p.reader.Close()
	p.L.Close()
}

// Start read/eval/print/loop
func (p *REPL) Start() {
	for {
		if str, err := p.loadline(); err != nil {
			log.Println(err)
			return
		}
		p.toolbox.exec(str)
	}
}

func incomplete(err error) bool {
	if lerr, ok := err.(*lua.ApiError); ok {
		if perr, ok := lerr.Cause.(*parse.Error); ok {
			return perr.Pos.Line == parse.EOF
		}
	}
	return false
}

func (p *REPL) loadline() (string, error) {
	p.reader.SetPrompt(PS1)
	if line, err := p.reader.Readline(); err == nil {
		if _, err := r.L.LoadString("return " + line); err == nil { // try add return <...> then compile
			return line, nil
		}
		return p.multiline(line)
	}
	return "", err
}

func (p *REPL) multiline(ml string) (string, error) {
	for {
		if _, err := r.L.LoadString(ml); err == nil { // try compile
			return ml, nil
		} else if !incomplete(err) { // syntax error, but not EOF
			return ml, nil
		} else { // read next line
			p.reader.SetPrompt(PS2)
			if line, err := p.reader.Readline(); err == nil {
				ml = ml + "\n" + line
			} else {
				return "", err
			}
		}
	}
}

func main() {
	r := NewREPL()
	r.Start()
	r.Close()
}
