package symbol

import (
	"fmt"
)

type Symbol struct {
	Name       string
	Attributes map[string]interface{}
	NameSpace  *Table
	Scope      *Table
	Id         uint
}

type Table struct {
	Parent *Table
	//Children []*Table //??? possibly needed for debugging
	symbols map[uint]*Symbol
	names   map[uint]string
	ids     map[string]uint
	nextId  uint
}

func (table *Table) NumSymbols() uint {
	return table.nextId
}

func (table *Table) ById(id uint) (*Symbol, bool) {
	sym, ok := table.symbols[id]
	if ok {
		return sym, true
	}
	if table.Parent != nil {
		return table.Parent.ById(id)
	}
	return nil, false
}

func (table *Table) ByName(name string) (*Symbol, bool) {
	id, ok := table.ids[name]
	if ok {
		return table.symbols[id], true
	}
	if table.Parent != nil {
		return table.Parent.ByName(name)
	}
	return nil, false
}

func (table *Table) ByQualifiedName(name []string) (*Symbol, bool) {
	if len(name) == 0 {
		return nil, false
	}
	t := table

	for len(name) > 0 {
		id, ok := t.ids[name[0]]
		if !ok {
			break
		}
		sym := t.symbols[id]
		if len(name) == 1 {
			return sym, true
		}
		name = name[1:]
		t = sym.NameSpace
	}
	if table.Parent != nil {
		return table.Parent.ByQualifiedName(name)
	}
	return nil, false
}

/*
func (table *Table) LocalSymbolById(id uint) (*Symbol, bool) {
	sym, ok := table.symbols[id]
	return sym, ok
}

func (table *Table) LocalSymbolByName(name string) (*Symbol, bool) {
	id, ok := table.ids[name]
	if ok {
		return table.symbols[id], true
	}
	return nil, false
}
*/

//Add a symbol with a scope
func (table *Table) Add(name string) (*Symbol, error) {
	_, exists := table.ids[name]
	if exists {
		return nil, fmt.Errorf("Name %q already defined in current scope")
	}

	ns := table.SubScope()
	id := table.nextId

	sym := &Symbol{
		Name:       name,
		Attributes: make(map[string]interface{}),
		NameSpace:  ns,
		Scope:      table,
		Id:         id,
	}

	table.symbols[id] = sym
	table.ids[name] = id
	table.names[id] = name

	table.nextId++

	return sym, nil
}

//Add a lower level scope
func (table *Table) SubScope() *Table {
	return &Table{
		Parent:  table,
		symbols: make(map[uint]*Symbol),
		names:   make(map[uint]string),
		ids:     make(map[string]uint),
		nextId:  0,
	}
}

//Create a global level symtab
func NewGlobalScope() *Table {
	return &Table{
		Parent:  nil,
		symbols: make(map[uint]*Symbol),
		names:   make(map[uint]string),
		ids:     make(map[string]uint),
		nextId:  0,
	}
}
