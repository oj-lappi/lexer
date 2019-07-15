package parse

import (
	"fmt"
	"kugg/compilers/lex"
	"strings"
)

//The value of a NodeType identifies the type of node.
type NodeType int

const (
	RootNode NodeType = -1 //RootNode is a special nodetype used to initialize the tree.
)

//NodeNames will be used whenever printing a Node of the given NodeType.
//
//Populate this if you want nice printouts.
var NodeNames = map[NodeType]string{
	RootNode: "root",
}

//String returns a string representation of the NodeType,
//which is either found in the NodeNames map or the underlying integer as a fallback.
//
//If you want the node to print nicely, you must populate the NodeNames map with names
//for the nodes you will be printing.
func (nt NodeType) String() string {
	s, ok := NodeNames[nt]
	if !ok {
		return fmt.Sprintf("%i", int(nt))
	}
	return s
}

//Node is a node in a parse tree
type Node interface {
	//Query type methods
	Type() NodeType
	Token() lex.Token
	Parent() Node
	Children() []Node
	Tree() *Tree
	FirstUncommittedAncestor(NodeType) Node

	Status() parseStatus
	IsTerminal() bool

	//Methods for manipulating tree structure
	AddChild(Node)
	AddChildren([]Node)
	ReplaceChild(Node, Node)
	RemoveChild(Node)
	ReplaceWith(Node)
	setParent(Node) //setParent is called by the parents AddChild method

	AddNonTerminal(NodeType, lex.Token) Node
	AddTerminal(NodeType, lex.Token) Node

	Commit()   //Commit sets parseStatus to FullyParsed
	RollBack() //RollBack removes all speculative nodes in a subtree

	//Methods for formatting
	PPrint(indent int) //PPrint pretty prints the parse tree in preorder
	String() string
}

type parseStatus bool

const (
	Speculative parseStatus = false
	FullyParsed parseStatus = true
)

//baseNode implements some common methods for all Nodes
type baseNode struct {
	typ         NodeType
	token       lex.Token
	parent      Node
	children    []Node
	tree        *Tree
	parseStatus parseStatus
	isTerminal  bool
}

//NewNonTerminal creates a new non-terminal node
func NewNonTerminal(typ NodeType, firstToken lex.Token, tree *Tree) Node {
	return &baseNode{
		typ:        typ,
		tree:       tree,
		token:      firstToken,
		children:   make([]Node, 0, 0),
		isTerminal: false}
}

//NewTerminal creates a new terminal node
func NewTerminal(typ NodeType, token lex.Token, tree *Tree) Node {
	return &baseNode{
		typ:        typ,
		tree:       tree,
		token:      token,
		isTerminal: true}

}

//panic because someone is trying to use a terminal node as a nonterminal node
func (n *baseNode) noChildren(action string) {
	n.tree.ErrorAtTokenf(n.token, "Can't %v children. This is a terminal node.", action)
}

//Type returns the NodeType of the Node
func (b *baseNode) Type() NodeType {
	return b.typ
}

//Children returns nil, since terminal nodes have no children
func (b *baseNode) Children() []Node {
	return b.children
}

//Parent returns the Nodes parent
func (b *baseNode) Parent() Node {
	return b.parent
}

//Tree returns the containing Tree
func (b *baseNode) Tree() *Tree {
	return b.tree
}

//FirstUncommittedAncestor gets the first ancestor of a specific type that has not been fully parsed yet.
func (b *baseNode) FirstUncommittedAncestor(typ NodeType) Node {

	p := b.Parent()

	for ; p != nil && p.Type() != typ && p.Status() == FullyParsed; p = p.Parent() {
	}

	return p
}

//Token returns the token corresponding to the node, or the leftmost one in case of a nonterminal
func (n *baseNode) Token() lex.Token {
	return n.token
}

//Status returns the parse status of the node, either FullyParsed or Speculative
func (b *baseNode) Status() parseStatus {
	return b.parseStatus
}

//IsTerminal checks if this is a terminal (leaf) node or not.
func (b *baseNode) IsTerminal() bool {
	return b.isTerminal
}

//String returns the NodeType turned into a string
func (n *baseNode) String() string {
	if n.isTerminal {
		return fmt.Sprintf("%v: %v", n.typ, n.token)
	}
	return fmt.Sprintf("%v", n.typ)
}

//AddChild panics, since terminal nodes have no children
func (n *baseNode) AddChild(nu Node) {
	if n.isTerminal {
		n.noChildren("add")
	} else {
		n.children = append(n.children, nu)
		n.setParent(n)
	}
}

//AddChild panics, since terminal nodes have no children
func (n *baseNode) AddChildren(ns []Node) {
	if n.isTerminal {
		n.noChildren("add")
	}
	n.children = append(n.children, ns...)
	for _, v := range ns {
		n.setParent(v)
	}
}

//AddNonbase panics, since terminal nodes have no children
func (n *baseNode) AddNonTerminal(typ NodeType, token lex.Token) Node {
	if n.isTerminal {
		n.noChildren("add")
	}
	nt := NewNonTerminal(typ, token, n.tree)
	n.AddChild(nt)
	return nt
}

//Addbase panics, since terminal nodes have no children
func (n *baseNode) AddTerminal(typ NodeType, token lex.Token) Node {
	if n.isTerminal {
		n.noChildren("add")
	}
	t := NewTerminal(typ, token, n.tree)
	n.AddChild(t)
	return t
}

//RemoveChild panics, since terminal nodes have no children
func (n *baseNode) RemoveChild(r Node) {
	if n.isTerminal {
		n.noChildren("remove")
	}

	for i, c := range n.children {
		if c == n {
			n.children = append(n.children[:i], n.children[i+1:]...)
			return
		}
	}
	n.tree.ErrorAtTokenf(n.Token(), "Cannot remove child %v. Node is not a child of %v.", n, n)

}

//ReplaceChild panics, since terminal nodes have no children
func (n *baseNode) ReplaceChild(old Node, nu Node) {
	if n.isTerminal {
		n.noChildren("replace")
	}

	for i, c := range n.children {
		if c == n {
			n.children[i] = nu
			return
		}
	}
	n.tree.ErrorAtTokenf(n.Token(), "Cannot replace child %v with %v. Node is not a child of %v.", old, nu, n)

}

//setParent is only used internally, call AddChild to link the two
func (n *baseNode) setParent(p Node) {
	n.parent = p
}

//ReplaceWith replaces the Node in the containing tree with another node
func (n *baseNode) ReplaceWith(nu Node) {
	n.parent.ReplaceChild(n, nu)
	n.parent = nil
}

//END tree manipulation methods

//Commit marks the node as fully parsed
func (b *baseNode) Commit() {
	b.parseStatus = FullyParsed
}

//RollBack removes all speculative nodes in a subtree
func (n *baseNode) RollBack() {
	//TODO:possible memory leak if child nodes still have references somewhere
	if n.parseStatus == Speculative {
		n.Parent().RemoveChild(n)
	}
	for _, child := range n.children {
		child.RollBack()
	}

}

//PPrint prints the tree indented, left to right
func (n *baseNode) PPrint(indent int) {
	fmt.Println(strings.Repeat("  ", indent), n)
	for _, v := range n.children {
		v.PPrint(indent + 1)
	}
}
