package parse

import (
	"fmt"
	"kugg/compilers/lex"

	"github.com/goccy/go-yaml"
)

func DeserializeTree(nodeNames map[NodeType]string, tokenNames map[lex.TokenType]string, text []byte) *Tree {

	m := make(map[string]interface{})
	err := yaml.Unmarshal(text, &m)

	if err != nil {
		panic(fmt.Sprintf("Error parsing serialized AST from YAML:%v\n", err))
	}

	nodeType := make(map[string]NodeType)
	for k, v := range nodeNames {
		nodeType[v] = k
	}

	tokType := make(map[string]lex.TokenType)
	for k, v := range tokenNames {
		tokType[v] = k
	}

	tree := NewTree("test_tree", string(text), nil)

	var walkYAMLtree func(map[string]interface{}) Node
	walkYAMLtree = func(yamlNode map[string]interface{}) Node {
		treeNode := baseNode{}
		hasToken := false
		var (
			nodeTokTyp lex.TokenType
			lexeme     string
		)
		for key, prop := range yamlNode {
			switch key {
			case "node":
				treeNode.typ = nodeType[prop.(string)]

			case "token":
				nodeTokTyp = tokType[prop.(string)]
				hasToken = true
			case "lexeme":
				lexeme = prop.(string)
			case "children":
				for _, child := range prop.([]interface{}) {
					switch child := child.(type) {
					case map[string]interface{}:
						treeNode.AddChild(walkYAMLtree(child))
					default:
						panic(fmt.Sprintf("Error walking deserialized YAML AST: expected children to be map[string]interface{}, but is %T:%v\n", child, child))
					}
				}

			}
		}
		if hasToken {
			treeNode.token = lex.DebugToken(nodeTokTyp, lexeme)
			//This is to get a pretty print with the tokens
			//Of course this means there are some print oddities, but that's fine
			treeNode.isTerminal = true
		}
		return &treeNode
	}

	tree.Root.AddChild(walkYAMLtree(m))

	return tree
}

//TODO: implement
func SerializeTree(nodeNames map[NodeType]string, tree Tree) string {
	return ""
}
