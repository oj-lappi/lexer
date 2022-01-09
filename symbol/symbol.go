package symbol

import (
	"fmt"
)

type Symbol struct {
	Name       string
	LocalId    uint
	GlobalId   uint
	Attributes map[string]interface{}
	NameSpace  *Table
	Scope      *Table
}

type Table struct {
	Parent   *Table
	Children []*Table

	symbols_by_name      map[string]*Symbol
	symbols_by_local_id  map[uint]*Symbol
	symbols_by_global_id map[uint]*Symbol
	nextId               uint
}

func (table *Table) NumSymbols() uint {
	return table.nextId
}

func (table *Table) ByLocalId(id uint) (*Symbol, bool) {
	sym, ok := table.symbols_by_local_id[id]
	if ok {
		return sym, true
	}
	return nil, false
}

func (table *Table) ByGlobalId(id uint) (*Symbol, bool) {
	sym, ok := table.symbols_by_global_id[id]
	if ok {
		return sym, true
	}
	if table.Parent != nil {
		return table.Parent.ByGlobalId(id)
	}
	return nil, false
}

func (table *Table) ByName(name string) (*Symbol, bool) {
	sym, ok := table.symbols_by_name[name]
	if ok {
		return sym, true
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
		sym, ok := t.symbols_by_name[name[0]]
		if !ok {
			break
		}
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

//Add a symbol with a scope
func (table *Table) Add(name string) (*Symbol, error) {
	_, exists := table.symbols_by_name[name]
	if exists {
		return nil, fmt.Errorf("Name %q already defined in current scope")
	}

	ns := table.SubScope()
	local_id := table.nextId

	sym := &Symbol{
		Name:       name,
		LocalId:    local_id,
		GlobalId:   0,
		Attributes: make(map[string]interface{}),
		NameSpace:  ns,
		Scope:      table,
	}

	table.symbols_by_name[name] = sym
	table.symbols_by_local_id[local_id] = sym

	table.nextId++

	return sym, nil
}

//Add a lower level scope
func (table *Table) SubScope() *Table {
	new_tab := &Table{
		Parent:               table,
		Children:             make([]*Table, 0),
		symbols_by_name:      make(map[string]*Symbol),
		symbols_by_local_id:  make(map[uint]*Symbol),
		symbols_by_global_id: make(map[uint]*Symbol),
		nextId:               0,
	}
	table.Children = append(table.Children, new_tab)
	return new_tab
}

//Recursively sets globally unique symbol ids for symbol tables
func (table *Table) ResolveGlobalIds() error {
	if table.Parent != nil {
		return nil
	}
	table.resolve_ids(0)

	//TODO: resolve symtabs not attached to symbols
	return nil

}

func (table *Table) resolve_ids(global_offset uint) uint {
	for i := uint(0); i < table.NumSymbols(); i++ {
		sym := table.symbols_by_local_id[i]
		//Set the global id
		sym.GlobalId = sym.LocalId + global_offset
		table.symbols_by_global_id[sym.GlobalId] = sym
		//Recurse
		sym.NameSpace.resolve_ids(global_offset + i + 1)
		global_offset += sym.NameSpace.nextId
	}
	return global_offset
}

//Create a global level symtab
func NewGlobalScope() *Table {
	return &Table{
		Parent:               nil,
		Children:             make([]*Table, 0),
		symbols_by_name:      make(map[string]*Symbol),
		symbols_by_local_id:  make(map[uint]*Symbol),
		symbols_by_global_id: make(map[uint]*Symbol),
		nextId:               0,
	}
}
