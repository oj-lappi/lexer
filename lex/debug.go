package lex

import (
	"fmt"
	"strconv"
	"strings"
)

func DebugToken(typ TokenType, value string) Token {
	return &token{typ: typ, value: value}
}

func DeserializeTokens(tokenNames map[TokenType]string, text string) []Token {

	tokType := make(map[string]TokenType)

	for k, v := range tokenNames {
		tokType[v] = k
	}

	var tokens []Token
	for _, line := range strings.Split(text, "\n") {
		var cols []string
		//Ignore empty lines
		if strings.Trim(line, " \t") == "" {
			continue
		}
		for _, col := range strings.Split(line, ":") {
			cols = append(cols, strings.Trim(col, " \t"))
		}

		var tok token
		switch len(cols) {
		case 2:
			tok.value = cols[1]
			fallthrough
		case 1:
			var ok bool
			tok.typ, ok = tokType[cols[0]]
			if !ok {
				//TODO: error out
				fmt.Printf("No token named \"%s\" in lookup table.\n", cols[0])
			}
		}
		tokens = append(tokens, &tok)
	}
	return tokens
}

func SerializeTokens(tokenNames map[TokenType]string, tokens []Token) string {
	var str string
	for _, tok := range tokens {
		name, ok := tokenNames[tok.Type()]
		if ok {
			str += name
		} else {
			str += strconv.Itoa(int(tok.Type()))
		}

		if tok.Lexeme() != "" {
			str += ":"
			str += tok.Lexeme()
		}
		str += "\n"
	}
	return str
}
