package utils

import (
	"github.com/KyberNetwork/promm-sdk-go/constants"
	"github.com/linhbkhn95/int256"
)

func AddDelta(x, y *int256.Int) *int256.Int {
	if y.Cmp(constants.Zero) < 0 {
		return int256.New().Sub(x, int256.New().Mul(y, constants.NegativeOne))
	} else {
		return int256.New().Add(x, y)
	}
}
