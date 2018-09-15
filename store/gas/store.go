package gas

import (
	"github.com/cosmos/cosmos-sdk/store/types"
)

var _ types.KVStore = &gasKVStore{}

// gasKVStore applies gas tracking to an underlying kvstore
type gasKVStore struct {
	tank   *types.GasTank
	parent types.KVStore
}

// nolint
func NewStore(tank *types.GasTank, parent types.KVStore) *gasKVStore {
	kvs := &gasKVStore{
		tank:   tank,
		parent: parent,
	}
	return kvs
}

// Implements types.KVStore.
func (gi *gasKVStore) Get(key []byte) (value []byte) {
	gi.tank.ReadFlat()
	value = gi.parent.Get(key)
	// TODO overflow-safe math?
	gi.tank.ReadBytes(len(value))

	return value
}

// Implements types.KVStore.
func (gi *gasKVStore) Set(key []byte, value []byte) {
	gi.tank.WriteFlat()
	// TODO overflow-safe math?
	gi.tank.WriteBytes(len(value))
	gi.parent.Set(key, value)
}

// Implements types.KVStore.
func (gi *gasKVStore) Has(key []byte) bool {
	gi.tank.HasFlat()
	return gi.parent.Has(key)
}

// Implements types.KVStore.
func (gi *gasKVStore) Delete(key []byte) {
	// No gas costs for deletion
	gi.parent.Delete(key)
}

// Implements types.KVStore.
func (gi *gasKVStore) Iterator(start, end []byte) types.Iterator {
	return gi.iterator(start, end, true)
}

// Implements types.KVStore.
func (gi *gasKVStore) ReverseIterator(start, end []byte) types.Iterator {
	return gi.iterator(start, end, false)
}

func (gi *gasKVStore) iterator(start, end []byte, ascending bool) types.Iterator {
	var parent types.Iterator
	if ascending {
		parent = gi.parent.Iterator(start, end)
	} else {
		parent = gi.parent.ReverseIterator(start, end)
	}
	return newGasIterator(gi.tank, parent)
}

type gasIterator struct {
	tank   *types.GasTank
	parent types.Iterator
}

func newGasIterator(tank *types.GasTank, parent types.Iterator) types.Iterator {
	return &gasIterator{
		tank:   tank,
		parent: parent,
	}
}

// Implements Iterator.
func (g *gasIterator) Domain() (start []byte, end []byte) {
	return g.parent.Domain()
}

// Implements Iterator.
func (g *gasIterator) Valid() bool {
	return g.parent.Valid()
}

// Implements Iterator.
func (g *gasIterator) Next() {
	g.parent.Next()
}

// Implements Iterator.
func (g *gasIterator) Key() (key []byte) {
	g.tank.KeyFlat()
	key = g.parent.Key()
	return key
}

// Implements Iterator.
func (g *gasIterator) Value() (value []byte) {
	value = g.parent.Value()
	g.tank.ValueFlat()
	g.tank.ValueBytes(len(value))
	return value
}

// Implements Iterator.
func (g *gasIterator) Close() {
	g.parent.Close()
}
