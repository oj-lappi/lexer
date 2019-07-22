package parse

import (
	"fmt"
	"kugg/compilers/lex"
)

type Tree struct {
	Root        Node        //Root of the parse tree
	Curr        Node        //Current node,
	Speculative Node        //Current subtree being parsed, will be attached to curr after a commmit
	Buffer      []lex.Token //Full token stream
	NameSpace   []string    //Current scope
	Pos         int         //Position of Current token in Buffer

	name       string
	text       string
	lexer      lex.Lexer
	startState StateFn //The StateFn first called
}

//StateFn is used to represent the state of the parser and the state transitions.
//
//Any StateFn returned by a StateFn is the new state and will be run next.
//Returning nil from a StateFn will stop the parser.
type StateFn func(*Tree) StateFn

func NewTree(name, text string, start StateFn) *Tree {
	tree := &Tree{
		name:       name,
		text:       text,
		startState: start,
		NameSpace:  make([]string, 0, 4),    //Three is the default amount of names expected (just a guess)
		Buffer:     make([]lex.Token, 0, 4), //Three is also the default length of the Buffer needed (more educated guess)
		Pos:        -1,
	}

	tree.Root = NewNonTerminal(RootNode, nil, tree)
	tree.Curr = tree.Root
	tree.Speculative = tree.Root

	return tree
}

//Parse runs the parser on the tokens produced by the lexer given
//
//Parse uses the state function pattern.
//This may change if another approach seems better.
func (tree *Tree) Parse(lexer lex.Lexer) (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v\n", r)
		}
	}()

	tree.lexer = lexer
	state := tree.startState
	for state != nil {
		state = state(tree)
	}
	return
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

/*
//Commit adds the speculative node to the current node
func (tree *Tree) Commit() {
	tree.Curr.AddChild(tree.Speculative)
	tree.ClearBuffer()
}
*/

//ClearBuffer clears the Buffer up until the position
//
//Use of this function is probably ill-advised unless you have huge token streams.
func (tree *Tree) ClearBuffer() {
	tree.Buffer = tree.Buffer[tree.Pos:]
	tree.Pos = -1
}

//AddNonTerminal adds a non-terminal node to the current subtree being built
func (tree *Tree) AddNonTerminal(typ NodeType, token lex.Token) Node {
	if tree.Speculative != nil {
		return tree.Speculative.AddNonTerminal(typ, token)
	}
	return nil
}

//AddTerminal adds a terminal to the current subtree being built
func (tree *Tree) AddTerminal(typ NodeType, token lex.Token) Node {
	if tree.Speculative != nil {
		return tree.Speculative.AddTerminal(typ, token)
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
	panic(fmt.Sprintf("%v,%v : ", token.Line(), token.Row()) + fmt.Sprintf(format, args...))
}

//Unexpected panics with a message of the form "expected x, got y" for Unexpected(y,x)
//
//The order is confusing. TODO:switch order
func (tree *Tree) Unexpected(unexpected interface{}, expected interface{}) {
	tree.Errorf("expected %v, got %v.", expected, unexpected)
}

//PPrint pretty prints (indents) the parse tree in preorder
func (tree *Tree) PPrint() {
	tree.Root.PPrint(0)
}
