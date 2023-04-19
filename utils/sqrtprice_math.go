package utils

import (
	"errors"

	"github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/linhbkhn95/int256"

	"github.com/KyberNetwork/promm-sdk-go/constants"
)

var (
	ErrSqrtPriceLessThanZero = errors.New("sqrt price less than zero")
	ErrLiquidityLessThanZero = errors.New("liquidity less than zero")
	ErrInvariant             = errors.New("invariant violation")
)
var MaxUint160 = int256.New().Sub(int256.New().Exp(constants.Two, int256.NewInt(160), nil), constants.One)

func multiplyIn256(x, y *int256.Int) *int256.Int {
	product := int256.New().Mul(x, y)
	return product.And(product, int256.MustFromBig(entities.MaxUint256))
}

func addIn256(x, y *int256.Int) *int256.Int {
	sum := int256.New().Add(x, y)
	return sum.And(sum, int256.MustFromBig(entities.MaxUint256))
}

func GetAmount0Delta(sqrtRatioAX96, sqrtRatioBX96, liquidity *int256.Int, roundUp bool) *int256.Int {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) >= 0 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}

	numerator1 := int256.New().Lsh(liquidity, 96)
	numerator2 := int256.New().Sub(sqrtRatioBX96, sqrtRatioAX96)

	if roundUp {
		return MulDivRoundingUp(MulDivRoundingUp(numerator1, numerator2, sqrtRatioBX96), constants.One, sqrtRatioAX96)
	}
	numerator1.Mul(numerator1, numerator2)
	return numerator1.Div(numerator1.Div(numerator1, sqrtRatioBX96), sqrtRatioAX96)
}

func GetAmount1Delta(sqrtRatioAX96, sqrtRatioBX96, liquidity *int256.Int, roundUp bool) *int256.Int {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) >= 0 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}

	if roundUp {
		return MulDivRoundingUp(liquidity, int256.New().Sub(sqrtRatioBX96, sqrtRatioAX96), constants.Q96)
	}
	temp := int256.New().Sub(sqrtRatioBX96, sqrtRatioAX96)
	return temp.Div(temp.Mul(liquidity, temp), constants.Q96)
}

func GetNextSqrtPriceFromInput(sqrtPX96, liquidity, amountIn *int256.Int, zeroForOne bool) (*int256.Int, error) {
	if sqrtPX96.Cmp(constants.Zero) <= 0 {
		return nil, ErrSqrtPriceLessThanZero
	}
	if liquidity.Cmp(constants.Zero) <= 0 {
		return nil, ErrLiquidityLessThanZero
	}
	if zeroForOne {
		return getNextSqrtPriceFromAmount0RoundingUp(sqrtPX96, liquidity, amountIn, true)
	}
	return getNextSqrtPriceFromAmount1RoundingDown(sqrtPX96, liquidity, amountIn, true)
}

func GetNextSqrtPriceFromOutput(sqrtPX96, liquidity, amountOut *int256.Int, zeroForOne bool) (*int256.Int, error) {
	if sqrtPX96.Cmp(constants.Zero) <= 0 {
		return nil, ErrSqrtPriceLessThanZero
	}
	if liquidity.Cmp(constants.Zero) <= 0 {
		return nil, ErrLiquidityLessThanZero
	}
	if zeroForOne {
		return getNextSqrtPriceFromAmount1RoundingDown(sqrtPX96, liquidity, amountOut, false)
	}
	return getNextSqrtPriceFromAmount0RoundingUp(sqrtPX96, liquidity, amountOut, false)
}

func getNextSqrtPriceFromAmount0RoundingUp(sqrtPX96, liquidity, amount *int256.Int, add bool) (*int256.Int, error) {
	if amount.Cmp(constants.Zero) == 0 {
		return sqrtPX96, nil
	}

	numerator1 := int256.New().Lsh(liquidity, 96)
	if add {
		product := multiplyIn256(amount, sqrtPX96)
		if int256.New().Div(product, amount).Cmp(sqrtPX96) == 0 {
			denominator := addIn256(numerator1, product)
			if denominator.Cmp(numerator1) >= 0 {
				return MulDivRoundingUp(numerator1, sqrtPX96, denominator), nil
			}
		}
		return MulDivRoundingUp(numerator1, constants.One, int256.New().Add(int256.New().Div(numerator1, sqrtPX96), amount)), nil
	} else {
		product := multiplyIn256(amount, sqrtPX96)
		if int256.New().Div(product, amount).Cmp(sqrtPX96) != 0 {
			return nil, ErrInvariant
		}
		if numerator1.Cmp(product) <= 0 {
			return nil, ErrInvariant
		}
		denominator := product.Sub(numerator1, product)
		return MulDivRoundingUp(numerator1, sqrtPX96, denominator), nil
	}
}

func getNextSqrtPriceFromAmount1RoundingDown(sqrtPX96, liquidity, amount *int256.Int, add bool) (*int256.Int, error) {
	if add {
		var quotient *int256.Int
		if amount.Cmp(MaxUint160) <= 0 {
			quotient = int256.New().Div(int256.New().Lsh(amount, 96), liquidity)
		} else {
			quotient = int256.New().Div(int256.New().Mul(amount, constants.Q96), liquidity)
		}
		return int256.New().Add(sqrtPX96, quotient), nil
	}

	quotient := MulDivRoundingUp(amount, constants.Q96, liquidity)
	if sqrtPX96.Cmp(quotient) <= 0 {
		return nil, ErrInvariant
	}
	return int256.New().Sub(sqrtPX96, quotient), nil
}
