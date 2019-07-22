package Gparse

import (
	"kugg/compilers/lex"
	"kugg/compilers/parse"
)

/*
This is an early draft of the main ideas behind this parser.

Intended as a design and possible tracer bullet.
*/

/*
Types for describing grammars

*/

var productionFns map[parse.NodeType]productionFn

type productionFn func(SententialForm) SententialForm

func buildMap(){
	for _,p := productions {
		productionFns[p.LHS] = productionFnFactory(p)
	}
}

func productionFnFactory(p *production) {
	return func(s SententialForm) SententialForm {
		//TODO:actually figure out the algorithm 
	}
}
