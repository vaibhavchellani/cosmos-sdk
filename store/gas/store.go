package gas

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.KVStore = &gasKVStore{}

// gasKVStore applies gas tracking to an underlying kvstore
type gasKVStore struct {
	tank   *sdk.GasTank
	parent sdk.KVStore
}

// nolint
func NewStore(tank *sdk.GasTank, parent sdk.KVStore) *gasKVStore {
	kvs := &gasKVStore{
		tank:   tank,
		parent: parent,
	}
	return kvs
}

// Implements sdk.KVStore.
func (gi *gasKVStore) Get(key []byte) (value []byte) {
	gi.tank.ReadFlat()
	value = gi.parent.Get(key)
	// TODO overflow-safe math?
	gi.tank.ReadBytes(len(value))

	return value
}

// Implements sdk.KVStore.
func (gi *gasKVStore) Set(key []byte, value []byte) {
	gi.tank.WriteFlat()
	// TODO overflow-safe math?
	gi.tank.WriteBytes(len(value))
	gi.parent.Set(key, value)
}

// Implements sdk.KVStore.
func (gi *gasKVStore) Has(key []byte) bool {
	gi.tank.HasFlat()
	return gi.parent.Has(key)
}

// Implements sdk.KVStore.
func (gi *gasKVStore) Delete(key []byte) {
	// No gas costs for deletion
	gi.parent.Delete(key)
}

// Implements sdk.KVStore.
func (gi *gasKVStore) Iterator(start, end []byte) sdk.Iterator {
	return gi.iterator(start, end, true)
}

// Implements sdk.KVStore.
func (gi *gasKVStore) ReverseIterator(start, end []byte) sdk.Iterator {
	return gi.iterator(start, end, false)
}

func (gi *gasKVStore) iterator(start, end []byte, ascending bool) sdk.Iterator {
	var parent sdk.Iterator
	if ascending {
		parent = gi.parent.Iterator(start, end)
	} else {
		parent = gi.parent.ReverseIterator(start, end)
	}
	return newGasIterator(gi.tank, parent)
}

type gasIterator struct {
	tank   *sdk.GasTank
	parent sdk.Iterator
}

func newGasIterator(tank *sdk.GasTank, parent sdk.Iterator) sdk.Iterator {
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
