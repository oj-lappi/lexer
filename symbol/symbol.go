package symbol

import "fmt"

type Table struct {
	Parent  *Table
	symbols map[uint]interface{}
	names   map[uint]string
	ids     map[string]uint
	nextId  uint
}

type Symbol struct {
	Name       string
	Attributes map[string]interface{}
	NameSpace  *Table
	Scope      *Table
	Id         uint
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

func (table *Table) Add(name string) (*Symbol, error) {
	_, exists := table.names[name]
	if exists {
		return nil, fmt.Errorf("Name %q already defined in current scope")
	}

	sym := Symbol{
		Name:       name,
		Attributes: make(map[string]interface{}),
		NameSpace:  NewGlobalTable(),
		Scope:      table,
		Id:         table.nextId,
	}
	table.nextId++
	return &sym, nil
}

func (table *Table) AddScope() *Table {
	return &Table{
		Parent:  table,
		symbols: make(map[uint]interface{}),
		names:   make(map[uint]string),
		ids:     make(map[string]uint),
		nextId:  0,
	}

}

func NewGlobalTable() *Table {
	return &Table{
		Parent:  nil,
		symbols: make(map[uint]interface{}),
		names:   make(map[uint]string),
		ids:     make(map[string]uint),
		nextId:  0,
	}
}
