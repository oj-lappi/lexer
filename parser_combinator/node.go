package lex

import (
	"fmt"
	"kugg/compilers/lex"
	"strings"
)

//TODO: There are two concepts getting muddied here:
//
//	 - The nodes as they will appear in the final parse tree
//	 - The production rules for nonterminals
//
//	Need to decide which one this is, or if they can use the same interface?

type NodeType int

const (
	RootNode NodeType = -1
)

//Node is a Syntactic lex.Token
type Node interface {
	//Query type methods
	Type() NodeType
	String() string
	Children() []Node
	Parent() Node
	Token() lex.Token
	Tree() *Tree
	IsTerminal() bool

	//Command type methods
	AddChild(Node)
	AddChildren([]Node)
	AddNonTerminal(NodeType) Node
	AddTerminal(NodeType, lex.Token) Node
	RemoveChild(Node)
	setParent(Node)

	//printing
	PPrint(indent int)
}

type baseNode struct {
	typ    NodeType
	parent Node
	tree   *Tree
}

//TerminalNode is used for the terminal nodes of the parse tree
type TerminalNode struct {
	*baseNode
	token lex.Token
}

//NonTerminalNode is used for non-terminal nodes of the parse tree
type NonTerminalNode struct {
	*baseNode
	children []Node
}

//NewNonTerminal creates a new non-terminal node
func NewNonTerminal(typ NodeType, tree *Tree) *NonTerminalNode {
	return &NonTerminalNode{
		baseNode: &baseNode{typ: typ, tree: tree},
		children: make([]Node, 0, 0),
	}
}

//NewTerminal creates a new terminal node
func NewTerminal(typ NodeType, token lex.Token, tree *Tree) *TerminalNode {
	return &TerminalNode{
		baseNode: &baseNode{typ: typ, tree: tree},
		token:    token,
	}

}

func (b *baseNode) Type() NodeType {
	return b.typ
}
func (b *baseNode) Parent() Node {
	return b.parent
}

func (b *baseNode) Tree() *Tree {
	return b.tree
}

//setParent is only used internally, call AddChild to link the two
func (b *baseNode) setParent(n Node) {
	b.parent = n
}

//NonTerminals and Terminals

//Children gets the children of the NonTerminal
func (nt *NonTerminalNode) Children() []Node {
	return nt.children
}

//AddChild adds the Node as a child, and then calls setParent on the node
func (nt *NonTerminalNode) AddChild(n Node) {
	nt.children = append(nt.children, n)
	n.setParent(nt)
}

//AddChildren adds multiple children
func (nt *NonTerminalNode) AddChildren(ns []Node) {
	nt.children = append(nt.children, ns...)
	for _, n := range ns {
		n.setParent(nt)
	}
}

//AddNonTerminal adds a non-terminal to the nodes children
//and returns the newly created node
func (nt *NonTerminalNode) AddNonTerminal(t NodeType) Node {
	n := NewNonTerminal(t, nt.tree)
	nt.AddChild(n)
	return n
}

//AddTerminal adds a terminal to the nodes children
//and returns the newly created node
func (nt *NonTerminalNode) AddTerminal(t NodeType, token lex.Token) Node {
	n := NewTerminal(t, token, nt.tree)
	nt.AddChild(n)
	return n
}

//RemoveChild removes a specific child
func (nt *NonTerminalNode) RemoveChild(n Node) {
	for i, c := range nt.children {
		if c == n {
			nt.children = append(nt.children[:i], nt.children[i+1:]...)
			return
		}
	}
}

//String returns the NodeType turned into a string
func (nt *NonTerminalNode) String() string {
	return fmt.Sprintf("%v", nt.typ)
}

//Token Returns nil
func (nt *NonTerminalNode) Token() lex.Token {
	return nil
}

//Isterminal returns false
func (nt *NonTerminalNode) IsTerminal() bool {
	return false
}

//PPrint prints the tree indented, left to right
func (nt *NonTerminalNode) PPrint(indent int) {
	fmt.Println(strings.Repeat(" ", indent), nt)
	for _, v := range nt.children {
		v.PPrint(indent + 1)
	}
}

//Destroy destroys the node and adds all chidren to the parent
func (nt *NonTerminalNode) Destroy() {
	nt.parent.AddChildren(nt.children)
	nt.parent.RemoveChild(nt)
	nt.parent = nil
	nt.children = nil
	nt.tree = nil
}

//Children returns nil, since terminal nodes have no children
func (t *TerminalNode) Children() []Node {
	return nil
}

//AddChild panics, since terminal nodes have no children
func (t *TerminalNode) AddChild(n Node) {
	panic("Can't add children to a terminal node")
}

//AddChild panics, since terminal nodes have no children
func (t *TerminalNode) AddChildren(ns []Node) {
	panic("Can't add children to a terminal node")
}

//AddNonTerminal panics, since terminal nodes have no children
func (t *TerminalNode) AddNonTerminal(typ NodeType) Node {
	panic("Can't add children to a terminal node")
}

//AddTerminal panics, since terminal nodes have no children
func (t *TerminalNode) AddTerminal(typ NodeType, token lex.Token) Node {
	panic("Can't add children to a terminal node")
}

//RemoveChild panics, since terminal nodes have no children
func (t *TerminalNode) RemoveChild(n Node) {
	panic("Trying to remove child node from terminal node, terminal nodes have no children")
}

//String returns the NodeType turned into a string
func (t *TerminalNode) String() string {
	return fmt.Sprintf("%v: %v", t.typ, t.token)
}

//Token returns the terminals token
func (t *TerminalNode) Token() lex.Token {
	return t.token
}

//IsTerminal returns true
func (t *TerminalNode) IsTerminal() bool {
	return true
}

//Destroy destroys the node by removing it from its parent
func (t *TerminalNode) Destroy() {
	t.parent.RemoveChild(t)
	t.parent = nil
	t.token = nil
	t.tree = nil
}

//PPrint prints the tree indented, left to right
func (t *TerminalNode) PPrint(indent int) {
	fmt.Println(strings.Repeat(" ", indent), t)
}
