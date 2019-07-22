package Gparse

/*
Type for describing a grammar
*/

type SymbolType int

type Symbol struct {
	SymbolType SymbolType
	terminal   bool
	lexeme     string
}

type SententialForm []SymbolType

type Production struct {
	LHS SymbolType
	RHS SententialForm
}
