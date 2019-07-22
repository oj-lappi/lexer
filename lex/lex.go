package lex

//This Lexer is influenced by RobPikes lecture on lexical analyzers for a templating language[0] and the corresponding source[1]

//[0]:	https://www.youtube.com/watch?v=HxaD_trXwRE
//[1]:	https://golang.org/src/text/template/parse/lex.go
//	https://golang.org/src/text/template/parse/node.go
//	https://golang.org/src/text/template/parse/parse.go

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

//Token is a token as scanned by the scanner
type Token interface {
	Type() TokenType
	Lexeme() string
	Pos() int
	Row() int
	Line() int
}

//token holds the token type, lexeme, and position as scanned by the scanner
type token struct {
	typ   TokenType //An integer identifying the token that has been scanned
	value string    //The scanned lexeme
	pos   int       //The position of the first rune in the source
	row   int       //The row of the first rune
	line  int       //The line of the first rune
}

func (t *token) String() string {
	switch {
	case t.typ == TokenEOF:
		return "EOF"
	case t.typ == TokenError:
		return fmt.Sprintf("%v:%v %v", t.line, t.row, t.value)
	case len(t.value) > 10:
		return fmt.Sprintf("%10q...", t.value)
	}
	return fmt.Sprintf("%q", t.value)
}

func (t *token) Type() TokenType {
	return t.typ
}
func (t *token) Lexeme() string {
	return t.value
}
func (t *token) Pos() int {
	return t.pos
}
func (t *token) Row() int {
	return t.row
}
func (t *token) Line() int {
	return t.line
}

//TokenType is an integer identifying a specific type of token
type TokenType int

//These special tokentypes are declared as negative ints.
//This is so that scanners can declare their own TokenTypes from iota
const (
	TokenError TokenType = -(1 + iota) //A scan error emitted by Errorf
	TokenEOF
)

const (
	Spaces = " \t\r\n"
)

const EOF = -1

//The Lexer interface provided here for convenience
type Lexer interface {
	Next() rune
	Peek() rune
	Back()
	Emit(t TokenType)
	Ignore()
	IgnoreCountNewLines()
	IgnoreSpaces()
	Errorf(format string, args ...interface{})
	Accept(valid string) bool
	AcceptRun(valid string) bool
	AcceptUntil(valid string) bool
	AcceptUnicodeRanges(ranges ...*unicode.RangeTable) bool
	AcceptUnicodeRangeRun(ranges ...*unicode.RangeTable) bool
	Switch(lookup map[rune]TokenType, fallback TokenType) TokenType
	Row() int
	NextToken() Token
	Drain()
	Run(StateFn)
}

//StateFn represents the state and the transitions to new states
//Using the StateFn pattern is hard with a reusable lexer.
//
//A trade off has been made. BaseLexer is exposed, and that is what
//a StateFn has as its argument. It's more convenient to handle a
//struct than to handle an interface in this case.
//
//If, after analysing a few implementations, I come to the conclusion that
//we are not manipulating the struct so much, even in harder grammars, then
//I will reduce the useful bits to an interface.
//
//But for now, you get a struct.
type StateFn func(*BaseLexer) StateFn

/*
//NewLexer returns a new Lexer with the given name and source
func NewLexer(name, source string) Lexer {
	return &BaseLexer{Name: name, Input: source}
}
*/

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

//Run starts the statemachine of the lexer
func (l *BaseLexer) run(state StateFn) {
	for state != nil {
		state = state(l)
	}
	close(l.Tokens)
}

func (l *BaseLexer) Run(state StateFn) {
	go l.run(state)
}

//Next returns the next rune in the source
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
	l.Tokens <- &token{
		typ:   t,
		value: l.Input[l.Start:l.Pos],
		pos:   l.Pos,
		line:  l.Line,
		row:   l.Row(),
	}
	l.Start = l.Pos
	l.StartLine = l.Line
}

//Ignore ignores the current token
func (l *BaseLexer) Ignore() {
	l.Start = l.Pos
	l.StartLine = l.Line
}

//IgnoreCountNewLines should be used if the scanner
//got to the current Pos without calling Next() or otherwise counting newlines
func (l *BaseLexer) IgnoreCountNewLines() {
	l.Line += strings.Count(l.Input[l.Start:l.Pos], "\n")
	l.Start = l.Pos
	l.StartLine = l.Line
}

//IgnoreSpaces will accept all space characters and then throw away the token
func (l *BaseLexer) IgnoreSpaces() {
	l.AcceptSpaces()
	l.Ignore()
}

//Errorf is used to emit a formatted error
func (l *BaseLexer) Errorf(format string, args ...interface{}) {
	//TODO: type switch to turn the runes in args into strings. They are being printed as char codes
	l.Tokens <- &token{
		typ:   TokenError,
		value: fmt.Sprintf(format, args...),
		pos:   l.Pos,
		line:  l.Line,
		row:   l.Row(),
	}
}

func (l *BaseLexer) Unexpected(unexpected interface{}, expected interface{}) {
	l.Errorf("expected %v, got %v .", expected, unexpected)
}

func (l *BaseLexer) NotAllowedInContext(notAllowed rune, context interface{}) {
	l.Errorf("rune %v not allowed in %v", notAllowed, context)
}

//Accept checks to see if the Next() rune is a part of the argument string.
//
//If no acceptable rune is found, it will not be consumed.
func (l *BaseLexer) Accept(valid string) bool {
	if strings.ContainsRune(valid, l.Next()) {
		return true
	}
	l.Back()
	return false
}

//AcceptRun calls Next() until it gets a rune which is not in the argument string
//
//It also reports whether a valid rune was found or not
func (l *BaseLexer) AcceptRun(valid string) (found bool) {
	for strings.ContainsRune(valid, l.Next()) {
		found = true
	}
	l.Back()
	return found
}

//AcceptUntil calls Next() until a rune from the argument string is found
//
//It also reports whether a rune was found or not
func (l *BaseLexer) AcceptUntil(until string) bool {
	r := l.Next()
	for r != EOF && !strings.ContainsRune(until, r) {
		r = l.Next()
	}
	l.Back()
	return r != EOF
}

//NOTE:In order to cut down on time specifying valid runes, an implementation
//can instead use RangeTable:s from the unicode package.
//
//The following convenience functions are provided for that purpose

//AcceptUnicodeRanges checks if the next rune is in any of the unicode ranges provided
func (l *BaseLexer) AcceptUnicodeRanges(ranges ...*unicode.RangeTable) bool {
	if unicode.In(l.Next(), ranges...) {
		return true
	}
	l.Back()
	return false
}

//AcceptUnicodeRangeRun will accept a run of runes from the unicode ranges provided
func (l *BaseLexer) AcceptUnicodeRangeRun(ranges ...*unicode.RangeTable) (found bool) {
	for unicode.In(l.Next(), ranges...) {
		found = true
	}
	l.Back()
	return found
}

//AcceptMatchOrRange will accept a rune that is either a part of the string given
//or in the unicode range tables given
func (l *BaseLexer) AcceptMatchOrRange(valid string, ranges ...*unicode.RangeTable) (found bool) {
	r := l.Next()
	if unicode.In(r, ranges...) || strings.ContainsRune(valid, r) {
		found = true
	}
	l.Back()
	return
}

//AcceptMatchOrRange will accept runes that is either a part of the string given
//or in the unicode range tables given
func (l *BaseLexer) AcceptMatchOrRangeRun(valid string, ranges ...*unicode.RangeTable) (found bool) {
	for r := l.Next(); unicode.In(r, ranges...) || strings.ContainsRune(valid, r); r = l.Next() {
		found = true
	}
	l.Back()
	return
}

//AcceptSpaces consumes all the space characters
func (l *BaseLexer) AcceptSpaces() (found bool) {
	for unicode.IsSpace(l.Next()) {
		found = true
	}
	l.Back()
	return
}

//Switch is a convenience function for multi-rune tokens using a lookup table
func (l *BaseLexer) Switch(lookup map[rune]TokenType, fallback TokenType) TokenType {
	v, ok := lookup[l.Peek()]
	if !ok {
		return fallback
	}
	return v
}

//Row finds the row of the first rune of the current token
func (l *BaseLexer) Row() int {
	lineStart := strings.LastIndex(l.Input[:l.Start], "\n")
	if lineStart >= 0 {
		return l.Start - lineStart
	}
	return l.Start
}

//NextToken returns the next token from the lexer synchronously
func (l *BaseLexer) NextToken() Token {
	return <-l.Tokens
}

//Drain empties the Tokens channel
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

func IsAlphaNumeric(r rune) bool {
	return unicode.In(r, unicode.Letter, unicode.Digit)
}
