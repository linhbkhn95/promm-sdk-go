package utils

import (
	"errors"

	"github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/linhbkhn95/int256"

	"github.com/KyberNetwork/promm-sdk-go/constants"
)

var ErrInvalidInput = errors.New("invalid input")
var powers = []int64{128, 64, 32, 16, 8, 4, 2, 1}
var powerBigInts = make([]*int256.Int, len(powers))

func init() {
	for i := range powers {
		powerBigInts[i] = int256.NewInt(powers[i])
	}
}

func MostSignificantBit(x *int256.Int) (int64, error) {
	if x.Cmp(constants.Zero) <= 0 {
		return 0, ErrInvalidInput
	}
	if x.Cmp(int256.MustFromBig(entities.MaxUint256)) > 0 {
		return 0, ErrInvalidInput
	}
	var msb int64
	for i, power := range powers {
		min := int256.New().Exp(constants.Two, powerBigInts[i], nil)
		if x.Cmp(min) >= 0 {
			x = int256.New().Rsh(x, uint(power))
			msb += power
		}
	}
	return msb, nil
}
