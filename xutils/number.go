// Package xutils is a utility package, it contains some utility functions.
package xutils

import (
	"log/slog"
	"math/big"
)

// BigIntE18ToFloat64 converts a big int to float64, the big int is in 1e18 scale
func BigIntE18ToFloat64(bi *big.Int, log *slog.Logger) float64 {
	bf := big.NewFloat(0).SetInt(bi)
	f, acc := bf.Quo(bf, big.NewFloat(1e18)).Float64()
	if acc != 0 {
		log.Error("cannot convert big int to float, has accuracy", "accuracy", acc, "int", bi.String())
		f = 0
	}
	return f
}
