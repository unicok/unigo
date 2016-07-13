package main

import (
	"bufio"
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gopkg.in/urfave/cli.v2"
)

const (
	tkType = iota
	tkName
	tkPayload
	tkColon
	tkString
	tkNumber
	tkEOF
	tkDesc
)

var (
	keywords = map[string]int{
		"packet_type": tkType,
		"name":        tkName,
		"payload":     tkPayload,
		"desc":        tkDesc,
	}
)

type token struct {
	typ     int
	literal string
	number  int
}

type apiExpr struct {
	PacketType int
	Name       string
	Payload    string
	Desc       string
}

var (
	tokenEOF   = &token{typ: tkEOF}
	tokenColon = &token{typ: tkColon}
)

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

func (p *Lexer) readDesc() string {
	var runes []rune
	for {
		r, _, err := p.reader.ReadRune()
		if err == io.EOF {
			break
		} else if r == '\r' {
			break
		} else if r == '\n' {
			p.lineno++
			break
		} else {
			runes = append(runes, r)
		}
	}
	return string(runes)
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

func (p *Lexer) next() *token {
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

	var runes []rune
	if unicode.IsLetter(r) {
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
		if tkid, ok := keywords[string(runes)]; ok {
			t.typ = tkid
		} else {
			t.typ = tkString
			t.literal = string(runes)
		}
		return t
	} else if unicode.IsNumber(r) {
		for {
			runes = append(runes, r)
			r, _, err = p.reader.ReadRune()
			if err == io.EOF {
				break
			} else if unicode.IsNumber(r) {
				continue
			} else {
				p.reader.UnreadRune()
				break
			}
		}
		t := &token{}
		t.typ = tkNumber
		n, _ := strconv.Atoi(string(runes))
		t.number = n
		return t
	} else if r == ':' {
		return tokenColon
	}

	log.Fatal("lex error @lien:", p.lineno)
	return nil
}

// Parser is
type Parser struct {
	exprs []apiExpr
	lexer *Lexer
}

func (p *Parser) init(lex *Lexer) {
	p.lexer = lex
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
	api := apiExpr{}

	p.match(tkType)
	p.match(tkColon)
	t := p.match(tkNumber)
	api.PacketType = t.number

	p.match(tkName)
	p.match(tkColon)
	t = p.match(tkString)
	api.Name = t.literal

	p.match(tkPayload)
	p.match(tkColon)
	t = p.match(tkString)
	api.Payload = t.literal

	p.match(tkDesc)
	p.match(tkColon)
	api.Desc = p.lexer.readDesc()

	p.exprs = append(p.exprs, api)
	return true
}

func main() {
	app := cli.App{
		Name:     "Protocol Handler Generator",
		Usage:    "handle api.txt",
		Compiled: time.Now(),
		Authors:  []*cli.Author{&cli.Author{Name: "Robin", Email: "amorwilliams@hotmail.com"}},
		Version:  "1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Value: "./api.txt", Usage: "input api.txt file"},
			&cli.IntFlag{Name: "min_proto", Aliases: []string{"min"}, Value: 0, Usage: "minimum proto number"},
			&cli.IntFlag{Name: "max_proto", Aliases: []string{"max"}, Value: 1000, Usage: "maximum proto number"},
			&cli.StringFlag{Name: "template", Aliases: []string{"t"}, Value: "./templates/server/api.tmpl", Usage: "template file"},
		},
		Action: func(c *cli.Context) error {
			// parse
			f, err := os.Open(c.String("file"))
			if err != nil {
				log.Fatal(err)
				return err
			}
			lex := Lexer{}
			lex.init(f)
			p := Parser{}
			p.init(&lex)
			for p.expr() {
			}

			// use template to generate fianl output
			funcMap := template.FuncMap{
				"isReq": func(api apiExpr) bool {
					if api.PacketType < c.Int("min_proto") || api.PacketType > c.Int("max_proto") {
						return false
					}
					if strings.HasSuffix(api.Name, "_req") {
						return true
					}
					return false
				},
			}
			tmpl, err := template.New("api.tmpl").Funcs(funcMap).ParseFiles(c.String("template"))
			if err != nil {
				log.Fatal(err)
				return err
			}
			err = tmpl.Execute(os.Stdout, p.exprs)
			if err != nil {
				log.Fatal(err)
				return err
			}

			return nil
		},
	}
	app.Run(os.Args)
}
