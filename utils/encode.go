package utils

import "github.com/linhbkhn95/int256"

/**
 * Returns the sqrt ratio as a Q64.96 corresponding to a given ratio of amount1 and amount0
 * @param amount1 The numerator amount i.e., the amount of token1
 * @param amount0 The denominator amount i.e., the amount of token0
 * @returns The sqrt ratio
 */
func EncodeSqrtRatioX96(amount1 *int256.Int, amount0 *int256.Int) *int256.Int {
	numerator := int256.New().Lsh(amount1, 192)
	denominator := amount0
	numerator.Div(numerator, denominator)
	return numerator.Sqrt(numerator)
}
