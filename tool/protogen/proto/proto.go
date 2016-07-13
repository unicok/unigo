package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"
	"unicode"

	"gopkg.in/urfave/cli.v2"
)

const (
	tkSymbol = iota
	tkStructBegin
	tkStructEnd
	tkDataType
	tkArray
	tkEOF
)

var (
	datatypes map[string]map[string]struct {
		T string `json:"t"` // type
		R string `json:"r"` // read
		W string `json:"w"` // write
	} // type -> language -> t/r/w
)

var (
	tokenEOF = &token{typ: tkEOF}
)

type (
	fieldInfo struct {
		Name  string
		Typ   string
		Array bool
	}
	structInfo struct {
		Name   string
		Fields []fieldInfo
	}
)

type token struct {
	typ     int
	literal string
	r       rune
}

func syntaxError(p *Parser) {
	log.Println("syntax error @line:", p.lexer.lineno)
	log.Println(">> \033[1;31m", p.lexer.lines[p.lexer.lineno-1], "\033[0m <<")
	os.Exit(-1)
}

// Lexer is
type Lexer struct {
	reader *bytes.Buffer
	lines  []string
	lineno int
}

func (p *Lexer) init(r io.Reader) {
	bts, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	// read each line
	scanner := bufio.NewScanner(bytes.NewBuffer(bts))
	for scanner.Scan() {
		p.lines = append(p.lines, scanner.Text())
	}

	// clear comment
	re := regexp.MustCompile("(?m:^#(.*)$)")
	bts = re.ReplaceAllLiteral(bts, nil)
	p.reader = bytes.NewBuffer(bts)
	p.lineno = 1
}

func (p *Lexer) next() (t *token) {
	defer func() {
		//log.Println(t)
	}()
	var r rune
	var err error
	for {
		r, _, err = p.reader.ReadRune()
		if err == io.EOF {
			return tokenEOF
		} else if unicode.IsSpace(r) {
			if r == '\n' {
				p.lineno++
			}
			continue
		}
		break
	}

	if r == '=' {
		for k := 0; k < 2; k++ { // check "==="
			r, _, err = p.reader.ReadRune()
			if err == io.EOF {
				return tokenEOF
			}
			if r != '=' {
				p.reader.UnreadRune()
				return &token{typ: tkStructBegin}
			}
		}
		return &token{typ: tkStructEnd}
	} else if unicode.IsLetter(r) {
		var runes []rune
		for {
			runes = append(runes, r)
			r, _, err = p.reader.ReadRune()
			if err == io.EOF {
				break
			} else if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
				continue
			} else {
				p.reader.UnreadRune()
				break
			}
		}

		t := &token{}
		t.literal = string(runes)
		if _, ok := datatypes[t.literal]; ok {
			t.typ = tkDataType
		} else if t.literal == "array" {
			t.typ = tkArray
		} else {
			t.typ = tkSymbol
		}

		return t
	}

	log.Fatal("lex error @line:", p.lineno)
	return nil
}

func (p *Lexer) eof() bool {
	for {
		r, _, err := p.reader.ReadRune()
		if err == io.EOF {
			return true
		} else if unicode.IsSpace(r) {
			if r == '\n' {
				p.lineno++
			}
			continue
		} else {
			p.reader.UnreadRune()
			return false
		}
	}
}

// Parser is
type Parser struct {
	lexer   *Lexer
	infos   []structInfo
	symbols map[string]bool
}

func (p *Parser) init(lex *Lexer) {
	p.lexer = lex
	p.symbols = make(map[string]bool)
}

func (p *Parser) match(typ int) *token {
	t := p.lexer.next()
	if t.typ != typ {
		syntaxError(p)
	}
	return t
}

func (p *Parser) expr() bool {
	if p.lexer.eof() {
		return false
	}
	info := structInfo{}

	t := p.match(tkSymbol)
	info.Name = t.literal
	p.symbols[t.literal] = true
	p.match(tkStructBegin)
	p.fields(&info)
	p.infos = append(p.infos, info)
	return true
}

func (p *Parser) fields(info *structInfo) {
	for {
		t := p.lexer.next()
		if t.typ == tkStructEnd {
			return
		}
		if t.typ != tkSymbol {
			syntaxError(p)
		}

		field := fieldInfo{Name: t.literal}
		t = p.lexer.next()
		if t.typ == tkArray {
			field.Array = true
			t = p.lexer.next()
		}

		if t.typ == tkDataType || t.typ == tkSymbol {
			field.Typ = t.literal
		} else {
			syntaxError(p)
		}

		info.Fields = append(info.Fields, field)
	}
}

func (p *Parser) semanticCheck() {
	for _, info := range p.infos {
	FIELDLOOP:
		for _, field := range info.Fields {
			if _, ok := datatypes[field.Typ]; !ok {
				if p.symbols[field.Typ] {
					continue FIELDLOOP
				}
				log.Fatal("symbol not found:", field)
			}
		}
	}
}

func main() {
	app := cli.App{
		Name:     "Protocol Handler Generator",
		Usage:    "handle api.txt",
		Compiled: time.Now(),
		Authors:  []*cli.Author{&cli.Author{Name: "Robin", Email: "amorwilliams@hotmail.com"}},
		Version:  "1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Value: "./proto.txt", Usage: "input proto.txt file"},
			&cli.StringFlag{Name: "binding", Aliases: []string{"b"}, Value: "go", Usage: `language type binding:"go","cs"`},
			&cli.StringFlag{Name: "template", Aliases: []string{"t"}, Value: "./templates/server/proto.tmpl", Usage: "template file"},
		},
		Action: func(c *cli.Context) error {
			// load primitives mapping
			f, err := os.Open("primitives.json")
			if err != nil {
				log.Fatal(err)
				return err
			}
			if err := json.NewDecoder(f).Decode(&datatypes); err != nil {
				log.Fatal(err)
				return err
			}

			// parse
			file, err := os.Open(c.String("file"))
			if err != nil {
				log.Fatal(err)
				return err
			}
			lexer := Lexer{}
			lexer.init(file)
			p := Parser{}
			p.init(&lexer)
			for p.expr() {
			}

			// semantic
			p.semanticCheck()

			// use template to generate final output
			funcMap := template.FuncMap{
				"Type": func(t string) string {
					return datatypes[t][c.String("binding")].T
				},
				"Read": func(t string) string {
					return datatypes[t][c.String("binding")].R
				},
				"Write": func(t string) string {
					return datatypes[t][c.String("binding")].W
				},
			}
			tmpl, err := template.New("proto.tmpl").Funcs(funcMap).ParseFiles(c.String("template"))
			if err != nil {
				log.Fatal(err)
				return err
			}
			err = tmpl.Execute(os.Stdout, p.infos)
			if err != nil {
				log.Fatal(err)
				return err
			}
			return nil
		},
	}
	app.Run(os.Args)
}
