package utils

import (
	cmn "github.com/tendermint/tendermint/libs/common"
)

func bz(s string) []byte { return []byte(s) }

func KeyFmt(i int) []byte { return bz(cmn.Fmt("key%0.8d", i)) }
func ValFmt(i int) []byte { return bz(cmn.Fmt("value%0.8d", i)) }
