package symbol

import (
	"fmt"
	"sort"
	"strings"
)

func (sym *Symbol) spprint(indent int) []string {
	ret := make([]string, 0, 4)
	indent_step := "  "
	out_prefix := strings.Repeat(indent_step, indent)
	in_prefix := strings.Repeat(indent_step, indent+1)
	ret = append(ret, fmt.Sprintf("%s{", out_prefix))
	ret = append(ret, fmt.Sprintf("%sname:%q", in_prefix, sym.Name))
	ret = append(ret, fmt.Sprintf("%slocal id:%d", in_prefix, sym.LocalId))
	ret = append(ret, fmt.Sprintf("%sglobal id:%d", in_prefix, sym.GlobalId))
	{ //Attributes
		no_attrs := true
		//Sort keys for micro_optimisation
		//Micro-optimisation: would like to reuse this array instead of allocating again and again
		attr_keys := make([]string, len(sym.Attributes))

		i := 0
		for k := range sym.Attributes {
			attr_keys[i] = k
			i++
		}
		sort.Strings(attr_keys)

		for _, k := range attr_keys {
			v := sym.Attributes[k]
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
	strs := table.SPPrint()
	fmt.Println(strings.Join(strs, "\n"))
}

func (table *Table) SPPrint() []string {
	return table.spprint(0)
}

func (table *Table) spprint(indent int) []string {
	ret := make([]string, 0)
	for id := uint(0); id < table.NumSymbols(); id++ {
		sym, ok := table.ByLocalId(id)
		if ok {
			ret = append(ret, sym.spprint(indent)...)
		}
		//else panic?
	}
	return ret
}
