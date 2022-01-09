package symbol

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

//The type AttributeFunction is used to pass in filter functions that transform YAML output
//into more domain-specific types
type AttributeFunction func(key string, val interface{}) interface{}

func DeserializeSymTab(text []byte, attribute_func AttributeFunction) *Table {

	m := make(map[string]interface{})
	err := yaml.Unmarshal(text, &m)

	if err != nil {
		panic(fmt.Sprintf("Error parsing serialized symtab from YAML:%v\n", err))
	}

	symtab := NewGlobalScope()

	var walkYAML func(map[string]interface{}, *Table)
	walkYAML = func(yamlSymbol map[string]interface{}, tab *Table) {

		local_id := tab.nextId

		sym := &Symbol{
			Name:       "",
			Attributes: make(map[string]interface{}),
			NameSpace:  tab.SubScope(),
			Scope:      tab,
			LocalId:    local_id,
			GlobalId:   0,
		}
		for key, prop := range yamlSymbol {
			switch key {
			case "symbol":
				sym.Name = prop.(string)
			case "attributes":
				for attr, attr_val := range prop.(map[string]interface{}) {
					sym.Attributes[attr] = attribute_func(attr, attr_val)
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
		tab.symbols_by_local_id[local_id] = sym
		tab.symbols_by_name[sym.Name] = sym
		tab.nextId++
	}

	walkYAML(m, symtab)
	symtab.ResolveGlobalIds()
	return symtab
}

func SerializeSymTab(symtab *Table) string {
	return ""
}
