package parse

import (
	"fmt"
	"kugg/compilers/lex"
)

type Tree struct {
	Root      Node        //Root of the parse tree
	Curr      Node        //Current node,
	Buffer    []lex.Token //Full token stream
	NameSpace []string    //Current scope
	Pos       int         //Position of Current token in Buffer
	NestLevel int         //How many nested expressions are there currently

	name      string
	text      string
	lexer     lex.Lexer
	parserFun ParseFn //the parsing entry point
}

type ParseFn func(*Tree)

func NewTree(name, text string, start ParseFn) *Tree {
	tree := &Tree{
		name:      name,
		text:      text,
		parserFun: start,
		NameSpace: make([]string, 0, 4),    //Three is the default amount of names expected (just a guess)
		Buffer:    make([]lex.Token, 0, 4), //Three is also the default length of the Buffer needed (more educated guess)
		Pos:       -1,
	}

	tree.Root = NewNonTerminal(RootNode, nil, tree)
	tree.Curr = tree.Root

	return tree
}

//Catches errors from the parser
func (tree *Tree) Parse(lexer lex.Lexer) (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v\n", r)
		}
	}()

	tree.lexer = lexer
	tree.parserFun(tree)
	return
}

//CurrentToken returns the token last received from the token stream
func (tree *Tree) CurrentToken() lex.Token {
	return tree.Buffer[tree.Pos]
}

//Next returns the next token from the lexer
func (tree *Tree) Next() lex.Token {
	if tree.Pos == len(tree.Buffer)-1 {
		tree.Buffer = append(tree.Buffer, tree.lexer.NextToken())
	}
	tree.Pos++
	return tree.Buffer[tree.Pos]
}

//Peek returns but does not consume the next token
func (tree *Tree) Peek() lex.Token {
	t := tree.Next()
	tree.Back()
	return t
}

//Back rewinds the position in the buffer by one
func (tree *Tree) Back() {
	tree.Pos--
}

//BackUntil goes back through the buffer until it finds a token of a certain type.
//Assumes the parser implementation knows that this token exists in the buffer.
func (tree *Tree) BackUntil(typ lex.TokenType) {
	for ; tree.Buffer[tree.Pos].Type() != typ; tree.Pos-- {
	}
}

//Commit adds the speculative node to the current node
func (tree *Tree) CommitSubTree() {
	if tree.Curr != nil {
		tree.Curr.CommitSubTree()
		tree.Curr = nil
	} else {
		panic("Trying to commit nonexisting node to tree")
	}
}

//Commit adds the speculative node to the current node
func (tree *Tree) Commit() {
	if tree.Curr != nil {
		tree.Curr.Commit()
		tree.Curr = nil
	} else {
		panic("Trying to commit nonexisting node to tree")
	}
}

//ClearBuffer clears the Buffer up until the position
//
//Use of this function is probably ill-advised unless you have huge token streams.
func (tree *Tree) ClearBuffer() {
	tree.Buffer = tree.Buffer[tree.Pos:]
	tree.Pos = -1
}

//AddNonTerminal adds a non-terminal node to the current subtree being built
func (tree *Tree) AddNonTerminal(typ NodeType, token lex.Token) Node {
	if tree.Curr != nil {
		return tree.Curr.AddNonTerminal(typ, token)
	}
	return nil
}

//AddTerminal adds a terminal to the current subtree being built
func (tree *Tree) AddTerminal(typ NodeType, token lex.Token) Node {
	if tree.Curr != nil {
		return tree.Curr.AddTerminal(typ, token)
	}
	return nil
}

//Errorf throws a formatted panic which will be caught in Parse
//if called from within a StateFn.
func (tree *Tree) Errorf(format string, args ...interface{}) {
	/*
		tree.Root = nil
		tree.Curr = nil
		tree.Buffer = nil
		tree.NameSpace = nil
	*/
	panic(fmt.Sprintf(format, args...))
}

//ErrorAtTokenf throws a formatted panic which will be caught in Parse
//if called from within a StateFn.
//
//The message will also print out the position in the source file which caused the error.
func (tree *Tree) ErrorAtTokenf(token lex.Token, format string, args ...interface{}) {
	panic(
		fmt.Sprintf("%v,%v : %s\n%s", token.Line(), token.Row(),
			fmt.Sprintf(format, args...),
			tree.lexer.TokenInContext(token)))
}

//Unexpected panics with a message of the form "expected x, got y" for Unexpected(y,x)
//
//The order is confusing. TODO:switch order
func (tree *Tree) Unexpected(unexpected interface{}, expected interface{}) {
	tok, ok := unexpected.(lex.Token)
	if ok {
		tree.ErrorAtTokenf(tok, "expected %v, got %v.", expected, unexpected)
	} else {
		tree.Errorf("expected %v, got %v.", expected, unexpected)
	}
}

//PPrint pretty prints (indents) the parse tree in preorder
func (tree *Tree) PPrint() {
	tree.Root.PPrint(0)
}
