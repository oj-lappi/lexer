package lex

//This Lexer is influenced by Rob Pikes lecture on lexical analyzers for a templating language[0] and the corresponding source[1]

//[0]:	https://www.youtube.com/watch?v=HxaD_trXwRE
//[1]:	https://golang.org/src/text/template/parse/lex.go
//	https://golang.org/src/text/template/parse/node.go
//	https://golang.org/src/text/template/parse/parse.go

import (
	"fmt"
	"kugg/stringwidth"
	"strings"
	"unicode"
	"unicode/utf8"
)

//TokenType is an integer identifying a specific type of token
type TokenType int

//These special tokentypes are declared as negative ints.
//This is so that scanners can declare their own TokenTypes from iota
const (
	LexingError TokenType = -(1 + iota) //A scan error emitted by Errorf
	EOF_Token
)

//TokenNames will be used whenever printing a Token of the given TokenType.
//
//Populate this if you want nice printouts.
var TokenNames = map[TokenType]string{
	LexingError: "LexingError",
	EOF_Token:   "EOF_Token",
}

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
	name := TokenNames[t.typ]
	switch {
	case t.typ == LexingError:
		return fmt.Sprintf("%v(%v:%v %v)", name, t.line, t.row, t.value)
	case len(t.value) > 10:
		return fmt.Sprintf("%v(%10q...)", name, t.value)
	default:
		return fmt.Sprintf("%v(%q)", name, t.value)
	}
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
	IgnoreSpaces()
	IgnoreNonNewlineSpaces()
	Errorf(format string, args ...interface{})
	Accept(valid string) bool
	AcceptSpaces() bool
	AcceptNonNewlineSpaces() bool
	AcceptRun(valid string) bool
	AcceptUntil(invalid string) bool
	AcceptUntilMatchOrRanges(invalid string, ranges ...*unicode.RangeTable) bool
	AcceptUnicodeRanges(ranges ...*unicode.RangeTable) bool
	AcceptUnicodeRangeRun(ranges ...*unicode.RangeTable) bool
	Switch(lookup map[rune]TokenType, fallback TokenType) TokenType
	Row() int
	NextToken() Token
	Drain()
	Run(StateFn)

	TokenInContext(Token) string //returns two lines of text:
	//The line in the source containing the token
	//, and a ^ cursor pointing at the start of the token
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
	Name   string
	Input  string
	Start  int //Start of current token
	Pos    int //Scanner position
	Width  int //Width of current rune
	Line   int //Scanner line position
	Tokens chan Token
	Lines  map[int]int //Line index -> position in source of first character on line
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
		l.Lines[l.Line] = l.Pos
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
		row:   l.Row(),
		line:  l.Line,
	}
	l.Start = l.Pos
}

//Ignore ignores the current token
func (l *BaseLexer) Ignore() {
	l.Start = l.Pos
}

//IgnoreSpaces will accept all space characters and then throw away the token
func (l *BaseLexer) IgnoreSpaces() {
	l.AcceptSpaces()
	l.Ignore()
}

//IgnoreSpaces will accept all non-newline space characters and then throw away the token
func (l *BaseLexer) IgnoreNonNewlineSpaces() {
	l.AcceptNonNewlineSpaces()
	l.Ignore()
}

//Errorf is used to emit a formatted error
func (l *BaseLexer) Errorf(format string, args ...interface{}) {
	//TODO: type switch to turn the runes in args into strings. They are being printed as char codes
	l.Tokens <- &token{
		typ:   LexingError,
		value: fmt.Sprintf(format, args...),
		pos:   l.Pos,
		line:  l.Line,
		row:   l.Row(),
	}
}

func (l *BaseLexer) UnexpectedRune(unexpected rune, expected interface{}) {
	l.Errorf("expected %v, got '%c'.", expected, unexpected)
}

func (l *BaseLexer) Unexpected(unexpected interface{}, expected interface{}) {
	l.Errorf("expected %v, got \"%v\".", expected, unexpected)
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

//AcceptUntilMatchOrRanges calls Next() until a rune from the argument string or in the ranges is found
//
//It also reports whether a rune was found or not
func (l *BaseLexer) AcceptUntilMatchOrRanges(until string, ranges ...*unicode.RangeTable) bool {
	r := l.Next()
	for r != EOF && !strings.ContainsRune(until, r) && !unicode.In(r, ranges...) {
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

//AcceptNonNewlineSpaces consumes all space characters except '\n' and '\r'
func (l *BaseLexer) AcceptNonNewlineSpaces() (found bool) {
	for IsNonNewlineSpace(l.Next()) {
		found = true
	}
	l.Back()
	return
}

//IsNonNewlineSpace returns false for '\r' and '\n', true for other white space characters, and false for other characters
func IsNonNewlineSpace(r rune) bool {
	if IsNewline(r) {
		return false
	}
	return unicode.IsSpace(r)
}

//IsNewline returns true if the rune is '\r' or '\n', false otherwise
func IsNewline(r rune) bool {
	switch r {
	case '\n', '\r':
		return true
	}
	return false
}

//Switch is a convenience function for multi-rune tokens using a lookup table
func (l *BaseLexer) Switch(lookup map[rune]TokenType, fallback TokenType) TokenType {
	v, ok := lookup[l.Peek()]
	if !ok {
		return fallback
	}
	return v
}

//TODO: clearly Row should be Column, right?
//Row finds the row of the first rune of the current token
func (l *BaseLexer) Row() int {
	ret := l.Start - l.Lines[l.Line]
	if ret < 0 {
		return 0
	}
	return ret
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

func (l *BaseLexer) TokenInContext(t Token) string {
	lineNum := t.Line()
	b := l.Lines[lineNum]
	e, ok := l.Lines[lineNum+1]
	line := ""
	if !ok {
		//race condition: The lexer hasn't gotten to the next line yet, so we find it
		//issue if token is a newline?
		o := strings.IndexRune(l.Input[b:], '\n')
		if o > 0 {
			line = l.Input[b : b+o]
		} else {
			line = l.Input[b:]
		}
	} else {
		line = l.Input[b : e-1]
	}
	//TODO:Boundserror here:
	//TODO:rip out stringwidth from codebase
	spaces := stringwidth.InSpaces(line[:t.Row()])
	return line + "\n" + strings.Repeat(" ", spaces-1) + "^"
}

//Lex creates a new scanner (BaseLexer) for an input string
func Lex(name string, input string) *BaseLexer {
	l := &BaseLexer{
		Name:   name,
		Input:  input,
		Tokens: make(chan Token),
		Line:   1,
		Lines:  map[int]int{},
	}
	return l
}

func IsAlphaNumeric(r rune) bool {
	return unicode.In(r, unicode.Letter, unicode.Digit)
}
