package gas

import (
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.KVStore = &gasKVStore{}

// gasKVStore applies gas tracking to an underlying kvstore
type gasKVStore struct {
	gasMeter  sdk.GasMeter
	gasConfig sdk.GasConfig
	parent    sdk.KVStore
}

// nolint
func NewGasKVStore(gasMeter sdk.GasMeter, gasConfig sdk.GasConfig, parent sdk.KVStore) *gasKVStore {
	kvs := &gasKVStore{
		gasMeter:  gasMeter,
		gasConfig: gasConfig,
		parent:    parent,
	}
	return kvs
}

// Implements sdk.KVStore.
func (gi *gasKVStore) Get(key []byte) (value []byte) {
	gi.gasMeter.ConsumeGas(gi.gasConfig.ReadCostFlat, "ReadFlat")
	value = gi.parent.Get(key)
	// TODO overflow-safe math?
	gi.gasMeter.ConsumeGas(gi.gasConfig.ReadCostPerByte*sdk.Gas(len(value)), "ReadPerByte")

	return value
}

// Implements sdk.KVStore.
func (gi *gasKVStore) Set(key []byte, value []byte) {
	gi.gasMeter.ConsumeGas(gi.gasConfig.WriteCostFlat, "WriteFlat")
	// TODO overflow-safe math?
	gi.gasMeter.ConsumeGas(gi.gasConfig.WriteCostPerByte*sdk.Gas(len(value)), "WritePerByte")
	gi.parent.Set(key, value)
}

// Implements sdk.KVStore.
func (gi *gasKVStore) Has(key []byte) bool {
	gi.gasMeter.ConsumeGas(gi.gasConfig.HasCost, "Has")
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

// Implements sdk.KVStore.
func (gi *gasKVStore) CacheWrap() sdk.CacheWrap {
	panic("cannot CacheWrap a GasKVStore")
}

// CacheWrapWithTrace implements the sdk.KVStore interface.
func (gi *gasKVStore) CacheWrapWithTrace(_ io.Writer, _ sdk.TraceContext) sdk.CacheWrap {
	panic("cannot CacheWrapWithTrace a GasKVStore")
}

func (gi *gasKVStore) iterator(start, end []byte, ascending bool) sdk.Iterator {
	var parent sdk.Iterator
	if ascending {
		parent = gi.parent.Iterator(start, end)
	} else {
		parent = gi.parent.ReverseIterator(start, end)
	}
	return newGasIterator(gi.gasMeter, gi.gasConfig, parent)
}

type gasIterator struct {
	gasMeter  sdk.GasMeter
	gasConfig sdk.GasConfig
	parent    sdk.Iterator
}

func newGasIterator(gasMeter sdk.GasMeter, gasConfig sdk.GasConfig, parent sdk.Iterator) sdk.Iterator {
	return &gasIterator{
		gasMeter:  gasMeter,
		gasConfig: gasConfig,
		parent:    parent,
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
	g.gasMeter.ConsumeGas(g.gasConfig.KeyCostFlat, "KeyFlat")
	key = g.parent.Key()
	return key
}

// Implements Iterator.
func (g *gasIterator) Value() (value []byte) {
	value = g.parent.Value()
	g.gasMeter.ConsumeGas(g.gasConfig.ValueCostFlat, "ValueFlat")
	g.gasMeter.ConsumeGas(g.gasConfig.ValueCostPerByte*sdk.Gas(len(value)), "ValuePerByte")
	return value
}

// Implements Iterator.
func (g *gasIterator) Close() {
	g.parent.Close()
}
