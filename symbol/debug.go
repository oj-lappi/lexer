package symbol

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

func DeserializeSymTab(text []byte) *Table {

	m := make(map[string]interface{})
	err := yaml.Unmarshal(text, &m)

	if err != nil {
		panic(fmt.Sprintf("Error parsing serialized symtab from YAML:%v\n", err))
	}

	symtab := NewGlobalScope()

	var walkYAML func(map[string]interface{}, *Table)
	walkYAML = func(yamlSymbol map[string]interface{}, tab *Table) {

		id := tab.nextId

		sym := &Symbol{
			Name:       "",
			Attributes: make(map[string]interface{}),
			NameSpace:  tab.SubScope(),
			Scope:      tab,
			Id:         id,
		}
		for key, prop := range yamlSymbol {
			switch key {
			case "symbol":
				sym.Name = prop.(string)
			case "attributes":
				for attr, attr_val := range prop.(map[string]interface{}) {
					sym.Attributes[attr] = attr_val
				}
			case "namespace":
				for _, child := range prop.([]interface{}) {
					switch child := child.(type) {
					case map[string]interface{}:
						walkYAML(child, sym.NameSpace)
					default:
						panic(fmt.Sprintf("Error walking deserialized YAML symtab: expected namespace of %s to be map[string]interface{}, but is %T:%v\n", sym.Name, child, child))
					}
				}
			}
		}

		if sym.Name == "" {
			panic(fmt.Sprintf("Unnamed symbol when deserializing symtab: %v\n", yamlSymbol))
		}
		tab.symbols[id] = sym
		tab.ids[sym.Name] = id
		tab.names[id] = sym.Name
		tab.nextId++
	}

	walkYAML(m, symtab)
	return symtab
}

func SerializeSymTab(symtab *Table) string {
	return ""
}
