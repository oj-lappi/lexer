package lex

import (
	"fmt"
	"kugg/compilers/lex"
)

type Tree struct {
	Root      Node        //Root of the parse tree
	Curr      Node        //Current node, can be used by external algorithms
	Buffer    []lex.Token //Full token stream
	NameSpace []string    //Current scope
	Pos       int         //Position of Current token in Buffer

	name       string
	text       string
	lexer      lex.Lexer
	startState StateFn //The StateFn first called
}

type StateFn func(*Tree) StateFn

func NewTree(name, text string, startState StateFn, firstTerminal NodeType) *Tree {
	tree := &Tree{
		name:       name,
		text:       text,
		startState: startState,
		NameSpace:  make([]string, 0, 3),    //Three is the default amount of names expected (just a guess)
		Buffer:     make([]lex.Token, 0, 3), //Three is also the default length of the Buffer needed (more educated guess)
		Pos:        -1,
		//reasoning: 3,6,12,24 is more lenient on memory than 4,8,16,32... or is it?
	}

	tree.Root = NewNonTerminal(RootNode, tree)
	tree.Curr = tree.Root

	return tree
}

func (tree *Tree) Parse(lexer lex.Lexer) {
	tree.lexer = lexer

	state := tree.startState
	for state != nil {
		state = state(tree)
	}
}

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

func (tree *Tree) Back() {
	tree.Pos--
}

//ClearBuffer clears the Buffer
func (tree *Tree) ClearBuffer() {
	tree.Buffer = tree.Buffer[:0]
	tree.Pos = -1
}

//AddNode adds a node to the tree and clears the Buffer
//Still feels clunky

func (tree *Tree) AddNonTerminal(typ NodeType) Node {
	return tree.Curr.AddNonTerminal(typ)
}

func (tree *Tree) AddTerminal(typ NodeType, token lex.Token) Node {
	return tree.Curr.AddTerminal(typ, token)
}

func (tree *Tree) Errorf(format string, args ...interface{}) {
	tree.Root = nil
	tree.Curr = nil
	tree.Buffer = nil
	tree.NameSpace = nil
	panic(fmt.Sprintf(format, args...))
}

func (tree *Tree) Unexpected(unexpected interface{}, expected interface{}) {
	tree.Errorf("expected %v, got %v.", expected, unexpected)
}

func (tree *Tree) PPrint() {
	tree.Root.PPrint(0)
}
