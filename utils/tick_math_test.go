package utils

import (
	"fmt"
	"testing"

	"github.com/linhbkhn95/int256"
	"github.com/stretchr/testify/assert"

	"github.com/KyberNetwork/promm-sdk-go/constants"
)

func TestGetSqrtRatioAtTick(t *testing.T) {
	_, err := GetSqrtRatioAtTick(MinTick - 1)
	assert.ErrorIs(t, err, ErrInvalidTick, "tick tool small")

	_, err = GetSqrtRatioAtTick(MaxTick + 1)
	assert.ErrorIs(t, err, ErrInvalidTick, "tick tool large")

	rmax, _ := GetSqrtRatioAtTick(MinTick)
	assert.Equal(t, rmax, MinSqrtRatio, "returns the correct value for min tick")

	r0, _ := GetSqrtRatioAtTick(0)
	assert.Equal(t, r0, int256.New().Lsh(constants.One, 96), "returns the correct value for tick 0")

	rmin, _ := GetSqrtRatioAtTick(MaxTick)
	assert.Equal(t, rmin, MaxSqrtRatio, "returns the correct value for max tick")
}

func TestGetTickAtSqrtRatio(t *testing.T) {
	tmin, _ := GetTickAtSqrtRatio(MinSqrtRatio)
	assert.Equal(t, tmin, MinTick, "returns the correct value for sqrt ratio at min tick")

	a := int256.New().Sub(MaxSqrtRatio, constants.One)
	fmt.Println("a", a.String())
	tmax, _ := GetTickAtSqrtRatio(a)
	assert.Equal(t, MaxTick-1, tmax, "returns the correct value for sqrt ratio at max tick")
}
