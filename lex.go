package lex

//This Lexer/parser influenced by RobPikes lecture on lexing[0] and the corresponding source[1]

//[0]:	https://www.youtube.com/watch?v=HxaD_trXwRE
//[1]:	https://golang.org/src/text/template/parse/lex.go
//	https://golang.org/src/text/template/parse/node.go
//	https://golang.org/src/text/template/parse/parse.go

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Token struct {
	Type  TokenType
	Value string
	Pos   int
	Row   int
	Line  int
}

func (t Token) String() string {
	switch {
	case t.Type == TokenEOF:
		return "EOF"
	case t.Type == TokenError:
		return t.Value
	case len(t.Value) > 10:
		return fmt.Sprintf("%10q...", t.Value)
	}
	return fmt.Sprintf("%q", t.Value)
}

type TokenType int

//These are tokens which will be used in any language
const (
	TokenError TokenType = -(1 + iota) //a lex error
	TokenEOF                           //EOF
)

//These TokenTypes are just common ones which can be reused
//although I'm not sure if it's worth it, unless it helps the
//parser (a reusable one)
const (
	TokenIdentifier   TokenType = iota //a name
	TokenOperator                      //=, !=, >, etc
	TokenParenStart                    //(
	TokenParenEnd                      //)
	TokenBracketStart                  //{
	TokenBracketEnd                    //}
	TokenComment                       //#....
)

const EOF = -1

var (
	Letters = "abcdefghijklmnopqrstuvwxyzåäöABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ"
	Digits  = "0123456789"
	Spaces  = " \t\r\n"
)

type Lexer interface {
	Next() rune
	Peek() rune
	Back()
	Emit(t TokenType)
	Ignore()
	IgnoreDontCountNewLines()
	Errorf(format string, args ...interface{})
	Accept(valid string) bool
	AcceptRun(valid string) bool
	AcceptUntil(valid string) bool
	Switch(lookup map[rune]TokenType, fallback TokenType) TokenType
	SwitchRun(lookup map[string]TokenType, fallback TokenType) TokenType
	Row() int
	NextToken() Token
	Drain()
}

//BaseLexer holds the State of the scanner
type BaseLexer struct {
	Name      string
	Input     string
	Start     int //Start of current token
	Pos       int //Scanner position
	Width     int //Width of current rune
	StartLine int //Start line of current token
	Line      int //Scanner line position
	Tokens    chan Token
	//err    error //saw this in a config lang implementation
}

//Next returns the next rune
func (l *BaseLexer) Next() rune {
	if l.Pos >= len(l.Input) {
		l.Width = 0
		return EOF
	}

	r, w := utf8.DecodeRuneInString(l.Input[l.Pos:])
	l.Width = w
	l.Pos += l.Width
	if r == '\n' {
		l.Line++
	}
	return r
}

//Peek returns but does not consume the next rune
func (l *BaseLexer) Peek() rune {
	r := l.Next()
	l.Back()
	return r
}

//Back goes back a rune
func (l *BaseLexer) Back() {
	l.Pos -= l.Width
	// Correct newline count.
	if l.Width == 1 && l.Input[l.Pos] == '\n' {
		l.Line--
	}
}

//Emit emits a token to the channel
func (l *BaseLexer) Emit(t TokenType) {
	l.Tokens <- Token{
		Type:  t,
		Value: l.Input[l.Start:l.Pos],
		Pos:   l.Pos,
		Line:  l.Line,
		Row:   l.Row(),
	}
	l.Start = l.Pos
	l.StartLine = l.Line
}

//Ignore slurps up the line til now
func (l *BaseLexer) Ignore() {
	//This is in Rob Pikes implementation but I don't really understand why yet
	l.Line += strings.Count(l.Input[l.Start:l.Pos], "\n")
	l.Start = l.Pos
	l.StartLine = l.Line
}

func (l *BaseLexer) IgnoreDontCountNewLines() {
	l.Start = l.Pos
	l.StartLine = l.Line
}

func (l *BaseLexer) Errorf(format string, args ...interface{}) {
	l.Tokens <- Token{
		Type:  TokenError,
		Value: fmt.Sprintf(format, args),
		Pos:   l.Pos,
		Line:  l.Line,
		Row:   l.Row(),
	}
}

func (l *BaseLexer) Accept(valid string) bool {
	if strings.ContainsRune(valid, l.Next()) {
		return true
	}
	l.Back()
	return false
}

func (l *BaseLexer) AcceptRun(valid string) bool {
	found := false
	for strings.ContainsRune(valid, l.Next()) {
		found = true
	}
	l.Back()
	return found
}

func (l *BaseLexer) AcceptUntil(valid string) bool {
	r := l.Next()
	for r != EOF && !strings.ContainsRune(valid, r) {
		r = l.Next()
	}
	l.Back()
	return r != EOF
}

func (l *BaseLexer) Switch(lookup map[rune]TokenType, fallback TokenType) TokenType {
	v, ok := lookup[l.Next()]
	if !ok {
		return fallback
	}
	return v
}

func (l *BaseLexer) SwitchRun(lookup map[string]TokenType, fallback TokenType) TokenType {
	for k, v := range lookup {
		if l.AcceptRun(k) {
			return v
		}
	}
	return fallback

}

//Finds the beginning row of the current token
func (l *BaseLexer) Row() int {
	lineStart := strings.LastIndex(l.Input[:l.Pos], "\n")
	if lineStart >= 0 {
		return l.Pos - lineStart
	}
	return l.Pos
}

func (l *BaseLexer) NextToken() Token {
	return <-l.Tokens
}

func (l *BaseLexer) Drain() {
	for range l.Tokens {
	}
}

//Lex creates a new scanner (BaseLexer) for an input string
func Lex(name string, input string) *BaseLexer {
	l := &BaseLexer{
		Name:      name,
		Input:     input,
		Tokens:    make(chan Token),
		Line:      1,
		StartLine: 1,
	}
	return l
}

//Convenience functions
func IsSpace(r rune) bool {
	return strings.ContainsRune(Spaces, r)
}

func IsLetter(r rune) bool {
	return strings.ContainsRune(Letters, r)
}

func IsDigit(r rune) bool {
	return strings.ContainsRune(Digits, r)
}

func IsAlphaNumeric(r rune) bool {
	return strings.ContainsRune(Letters+Digits, r)
}
