package lex

/*
import (
	"strconv"
	"strings"
)

func DeserializeTokenStream(tokenNames map[TokenType]string, text string) []Token {

	var tokType map[string]TokenType

	for k, v := range tokenNames {
		tokType[v] = k
	}

	var tokens []Token
	for line := range strings.Split(text, "\n") {
		var cols []string
		for col := range strings.Split(line, ":") {
			cols.append(col.Trim(" \t"))
		}

		var tok token
		switch cols.len {
		case 2:
			tok.value = cols[1]
		case 1:
			tok.typ = tokType[cols[0]]
		}
		tokens.append(tok)
	}
	return tokens
}

func SerializeTokenStream(tokenNames map[TokenType]string, tokens []Token) string {
	var str string
	for tok := range tokens {
		name, ok := tokenNames[tok.Type()]
		if ok {
			str += name
		} else {
			str += strconv.Itoa(tok.Type())
		}

		if tok.Lexeme() != "" {
			str += ":"
			str += tok.Lexeme()
		}
		str += "\n"
	}
	return str
}
*/
