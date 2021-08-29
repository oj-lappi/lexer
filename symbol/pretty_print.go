package symbol

import (
	"fmt"
	"strings"
)

//TODO: duplication... maybe simplify? or make pprint use spprint
func (sym *Symbol) pprint(indent int) {
	indent_step := "  "
	out_prefix := strings.Repeat(indent_step, indent)
	in_prefix := strings.Repeat(indent_step, indent+1)

	fmt.Printf("%s{\n%sname:%q\n", out_prefix, in_prefix, sym.Name)
	fmt.Printf("%sid:%d\n", in_prefix, sym.Id)

	{ //Print attributes
		no_attrs := true
		for k, v := range sym.Attributes {
			if no_attrs {
				fmt.Printf("%sattributes:\n%s(\n", in_prefix, in_prefix)
				no_attrs = false
			}
			fmt.Printf("%s%s%q:%v\n", in_prefix, indent_step, k, v)
		}
		if !no_attrs {
			fmt.Printf("%s)\n", in_prefix)
		}
	}
	if sym.NameSpace.nextId != 0 {
		fmt.Printf("%snamespace:\n%s[\n", in_prefix, in_prefix)
		sym.NameSpace.pprint(indent + 2)
		fmt.Printf("%s]\n", in_prefix)
	}
	fmt.Printf("%s},\n", out_prefix)
}

func (sym *Symbol) spprint(indent int) []string {
	ret := make([]string, 0, 4)
	indent_step := "  "
	out_prefix := strings.Repeat(indent_step, indent)
	in_prefix := strings.Repeat(indent_step, indent+1)
	ret = append(ret, fmt.Sprintf("%s{", out_prefix))
	ret = append(ret, fmt.Sprintf("%sname:%q", in_prefix, sym.Name))
	ret = append(ret, fmt.Sprintf("%sid:%d", in_prefix, sym.Id))
	{ //Attributes
		no_attrs := true
		for k, v := range sym.Attributes {
			if no_attrs {
				ret = append(ret, fmt.Sprintf("%sattributes:", in_prefix))
				ret = append(ret, fmt.Sprintf("%s(", in_prefix))
				no_attrs = false
			}
			//TODO: list types of e.g. strings will look wrong.
			ret = append(ret, fmt.Sprintf("%s%s%q:%v", in_prefix, indent_step, k, v))
		}
		if !no_attrs {
			ret = append(ret, fmt.Sprintf("%s)", in_prefix))
		}
	}
	if sym.NameSpace.nextId != 0 {
		ret = append(ret, fmt.Sprintf("%snamespace:", in_prefix))
		ret = append(ret, fmt.Sprintf("%s[", in_prefix))
		ret = append(ret, sym.NameSpace.spprint(indent+2)...)
		ret = append(ret, fmt.Sprintf("%s]", in_prefix))
	}
	ret = append(ret, fmt.Sprintf("%s},", out_prefix))
	return ret
}

//Pretty print symtab
func (table *Table) PPrint() {
	table.pprint(0)
}

//Pretty print helper
func (table *Table) pprint(indent int) {
	for _, sym := range table.symbols {
		sym.pprint(indent)
	}
}

func (table *Table) SPPrint() []string {
	return table.spprint(0)
}

func (table *Table) spprint(indent int) []string {
	ret := make([]string, 0)
	for _, sym := range table.symbols {
		ret = append(ret, sym.spprint(indent)...)
	}
	return ret
}
