package utils

import (
	"github.com/KyberNetwork/promm-sdk-go/constants"
	"github.com/linhbkhn95/int256"
)

func MulDivRoundingUp(a, b, denominator *int256.Int) *int256.Int {
	product := int256.New().Mul(a, b)
	result := int256.New().Div(product, denominator)
	if int256.New().Rem(product, denominator).Cmp(constants.Zero) != 0 {
		result.Add(result, constants.One)
	}
	return result
}

func MulDivRoundingDown(a, b, denominator *int256.Int) *int256.Int {
	product := int256.New().Mul(a, b)
	return product.Quo(product, denominator)
}

func MulDiv(a, b, denominator *int256.Int) *int256.Int {
	product := int256.New().Mul(a, b)
	return product.Div(product, denominator)
}

func GetSmallerRootOfQuadEqn(a, b, c *int256.Int) *int256.Int {
	// smallerRoot = (b - sqrt(b * b - a * c)) / a;
	tmp1 := int256.New().Mul(b, b)
	tmp2 := int256.New().Mul(a, c)
	tmp3 := tmp1.Sqrt(tmp1.Sub(tmp1, tmp2))
	tmp3.Sub(b, tmp3)
	return tmp3.Div(tmp3, a)
}
