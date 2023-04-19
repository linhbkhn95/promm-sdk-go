package utils

import (
	"github.com/KyberNetwork/promm-sdk-go/constants"
	"github.com/linhbkhn95/int256"
)

var FeeUnits = int256.New().Exp(int256.NewInt(10), int256.NewInt(5), nil)
var TwoFeeUnits = int256.New().Mul(FeeUnits, int256.NewInt(2))

// ComputeSwapStep computes the actual swap input / output amounts to be deducted or added,
// the swap fee to be collected and the resulting sqrtP
func ComputeSwapStep(
	sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, amountRemaining *int256.Int, feeInUnits constants.FeeAmount,
	exactIn, isToken0 bool,
) (sqrtRatioNextX96, amountIn, amountOut, deltaL *int256.Int, err error) {
	// in the event currentSqrtP == targetSqrtP because of tick movements, return
	// e.g. swapped up tick where specified price limit is on an initialised tick
	// then swapping down tick will cause next tick to be the same as the current tick
	if sqrtRatioCurrentX96.Cmp(sqrtRatioTargetX96) == 0 {
		return sqrtRatioCurrentX96, constants.Zero, constants.Zero, constants.Zero, nil
	}

	sqrtRatioNextX96 = int256.NewInt(0)

	usedAmount := calcReachAmount(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, feeInUnits, exactIn, isToken0)

	if exactIn && usedAmount.Cmp(amountRemaining) >= 0 || (!exactIn && usedAmount.Cmp(amountRemaining) <= 0) {
		usedAmount = amountRemaining
	} else {
		sqrtRatioNextX96 = sqrtRatioTargetX96
	}

	// ??? not use
	amountIn = usedAmount

	var absUsedAmount *int256.Int

	if usedAmount.Cmp(constants.Zero) >= 0 {
		absUsedAmount = usedAmount
	} else {
		absUsedAmount = int256.New().Mul(usedAmount, constants.NegativeOne)
	}

	if sqrtRatioNextX96.Cmp(constants.Zero) == 0 {
		deltaL = estimateIncrementalLiquidity(
			absUsedAmount, liquidity, sqrtRatioCurrentX96, feeInUnits, exactIn, isToken0,
		)

		sqrtRatioNextX96 = calcFinalPrice(absUsedAmount, liquidity, deltaL, sqrtRatioCurrentX96, exactIn, isToken0)
	} else {
		deltaL = calcIncrementalLiquidity(
			sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, absUsedAmount, exactIn, isToken0,
		)
	}

	amountOut = calcReturnedAmount(sqrtRatioCurrentX96, sqrtRatioNextX96, liquidity, deltaL, exactIn, isToken0)

	return
}

// calcReachAmount calculates the amount needed to reach targetSqrtP from currentSqrtP
// we cast currentSqrtP and targetSqrtP to uint256 as they are multiplied by TWO_FEE_UNITS or feeInFeeUnits
func calcReachAmount(
	sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity *int256.Int, feeInUnits constants.FeeAmount,
	exactIn, isToken0 bool,
) (reachAmount *int256.Int) {
	var absPriceDiff *int256.Int

	if sqrtRatioCurrentX96.Cmp(sqrtRatioTargetX96) >= 0 {
		absPriceDiff = int256.New().Sub(sqrtRatioCurrentX96, sqrtRatioTargetX96)
	} else {
		absPriceDiff = int256.New().Sub(sqrtRatioTargetX96, sqrtRatioCurrentX96)
	}

	if exactIn {
		// we round down so that we avoid taking giving away too much for the specified input
		// i.e. require less input qty to move ticks
		if isToken0 {
			// exactInput + swap 0 -> 1
			// numerator = 2 * liquidity * absPriceDiff
			// denominator = currentSqrtP * (2 * targetSqrtP - currentSqrtP * feeInFeeUnits / FEE_UNITS)
			// overflow should not happen because the absPriceDiff is capped to ~5%
			temp := int256.New().Mul(TwoFeeUnits, sqrtRatioTargetX96)
			denominator := int256.New().Sub(
				temp,
				int256.New().Mul(int256.NewInt(int64(feeInUnits)), sqrtRatioCurrentX96),
			)
			numerator := MulDiv(liquidity, temp.Mul(TwoFeeUnits, absPriceDiff), denominator)

			reachAmount = MulDiv(numerator, constants.Q96, sqrtRatioCurrentX96)
		} else {
			// exactInput + swap 1 -> 0
			// numerator: liquidity * absPriceDiff * (TWO_FEE_UNITS * targetSqrtP - feeInFeeUnits * (targetSqrtP + currentSqrtP))
			// denominator: (TWO_FEE_UNITS * targetSqrtP - feeInFeeUnits * currentSqrtP)
			// overflow should not happen because the absPriceDiff is capped to ~5%
			temp := int256.New().Mul(TwoFeeUnits, sqrtRatioCurrentX96)

			denominator := int256.New().Sub(
				temp,
				int256.New().Mul(int256.NewInt(int64(feeInUnits)), sqrtRatioTargetX96),
			)
			numerator := MulDiv(liquidity, temp.Mul(TwoFeeUnits, absPriceDiff), denominator)

			reachAmount = MulDiv(numerator, sqrtRatioCurrentX96, constants.Q96)
		}
	} else {
		// we will perform negation as the last step
		// we round down so that we require less output qty to move ticks
		if isToken0 {
			// exactOut + swap 0 -> 1
			// numerator: (liquidity)(absPriceDiff)(2 * currentSqrtP - deltaL * (currentSqrtP + targetSqrtP))
			// denominator: (currentSqrtP * targetSqrtP) * (2 * currentSqrtP - deltaL * targetSqrtP)
			// overflow should not happen because the absPriceDiff is capped to ~5%

			temp := int256.New().Mul(TwoFeeUnits, sqrtRatioCurrentX96)

			denominator := int256.New().Sub(
				temp,
				int256.New().Mul(int256.NewInt(int64(feeInUnits)), sqrtRatioTargetX96),
			)
			numerator := int256.New().Sub(
				denominator, temp.Mul(int256.NewInt(int64(feeInUnits)), sqrtRatioCurrentX96),
			)
			numerator = MulDiv(temp.Lsh(liquidity, 96), numerator, denominator)

			reachAmount = int256.New().Div(MulDiv(numerator, absPriceDiff, sqrtRatioCurrentX96), sqrtRatioTargetX96)
			reachAmount = temp.Mul(reachAmount, constants.NegativeOne)
		} else {
			// exactOut + swap 1 -> 0
			// numerator: liquidity * absPriceDiff * (TWO_FEE_UNITS * targetSqrtP - feeInFeeUnits * (targetSqrtP + currentSqrtP))
			// denominator: (TWO_FEE_UNITS * targetSqrtP - feeInFeeUnits * currentSqrtP)
			// overflow should not happen because the absPriceDiff is capped to ~5%
			temp := int256.New().Mul(TwoFeeUnits, sqrtRatioTargetX96)

			denominator := int256.New().Sub(
				temp,
				int256.New().Mul(int256.NewInt(int64(feeInUnits)), sqrtRatioCurrentX96),
			)
			numerator := int256.New().Sub(
				denominator, temp.Mul(int256.NewInt(int64(feeInUnits)), sqrtRatioTargetX96),
			)
			numerator = MulDiv(liquidity, numerator, denominator)

			reachAmount = MulDiv(numerator, absPriceDiff, constants.Q96)
			reachAmount = temp.Mul(reachAmount, constants.NegativeOne)
		}
	}

	return reachAmount
}

// calcReturnedAmount calculates returned output | input tokens in exchange for specified amount
// round down when calculating returned output (isExactInput) so we avoid sending too much
// round up when calculating returned input (!isExactInput) so we get desired output amount
func calcReturnedAmount(
	sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, deltaL *int256.Int, exactIn, isToken0 bool,
) (returnedAmount *int256.Int) {
	if isToken0 {
		if exactIn {
			// minimise actual output (<0, make less negative) so we avoid sending too much
			// returnedAmount = deltaL * nextSqrtP - liquidity * (currentSqrtP - nextSqrtP)
			returnedAmount = int256.New().Add(
				MulDivRoundingUp(deltaL, sqrtRatioTargetX96, constants.Q96),
				int256.New().Mul(
					MulDiv(
						liquidity, int256.New().Sub(sqrtRatioCurrentX96, sqrtRatioTargetX96), constants.Q96,
					), constants.NegativeOne,
				),
			)
		} else {
			// maximise actual input (>0) so we get desired output amount
			// returnedAmount = deltaL * nextSqrtP + liquidity * (nextSqrtP - currentSqrtP)
			returnedAmount = int256.New().Add(
				MulDivRoundingUp(deltaL, sqrtRatioTargetX96, constants.Q96),
				MulDivRoundingUp(liquidity, int256.New().Sub(sqrtRatioTargetX96, sqrtRatioCurrentX96), constants.Q96),
			)
		}
	} else {
		// returnedAmount = (liquidity + deltaL)/nextSqrtP - (liquidity)/currentSqrtP
		// if exactInput, minimise actual output (<0, make less negative) so we avoid sending too much
		// if exactOutput, maximise actual input (>0) so we get desired output amount
		temp := int256.New().Add(liquidity, deltaL)
		returnedAmount = temp.Add(
			MulDivRoundingUp(temp, constants.Q96, sqrtRatioTargetX96),
			int256.New().Mul(MulDivRoundingUp(liquidity, constants.Q96, sqrtRatioCurrentX96), constants.NegativeOne),
		)
	}

	if exactIn && returnedAmount.Cmp(constants.One) == 0 {
		// rounding make returnedAmount == 1
		returnedAmount = constants.Zero
	}

	return returnedAmount
}

// calcIncrementalLiquidity calculates deltaL, the swap fee to be collected for an intermediate swap step,
// where the next (temporary) tick will be crossed
func calcIncrementalLiquidity(
	sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, absAmount *int256.Int, exactIn, isToken0 bool,
) (deltaL *int256.Int) {

	// this is when we reach the target, then we have target_X96
	if isToken0 {
		// deltaL = nextSqrtP * (liquidity / currentSqrtP +/- absDelta)) - liquidity
		// needs to be minimum
		tmp := MulDiv(liquidity, constants.Q96, sqrtRatioCurrentX96)
		if exactIn {
			tmp.Add(tmp, absAmount)
		} else {
			tmp.Sub(tmp, absAmount)
		}
		tmp = MulDiv(sqrtRatioTargetX96, tmp, constants.Q96)

		// in edge cases where liquidity or absDelta is small
		// liquidity might be greater than nextSqrtP * ((liquidity / currentSqrtP) +/- absDelta))
		// due to rounding
		if tmp.Cmp(liquidity) > 0 {
			deltaL = tmp.Sub(tmp, liquidity)
		} else {
			deltaL = constants.Zero
		}
	} else {
		// deltaL = (liquidity * currentSqrtP +/- absDelta) / nextSqrtP - liquidity
		// needs to be minimum
		tmp := MulDiv(liquidity, sqrtRatioCurrentX96, constants.Q96)
		if exactIn {
			tmp.Add(tmp, absAmount)
		} else {
			tmp.Sub(tmp, absAmount)
		}
		tmp = MulDiv(tmp, constants.Q96, sqrtRatioTargetX96)

		// in edge cases where liquidity or absDelta is small
		// liquidity might be greater than nextSqrtP * ((liquidity / currentSqrtP) +/- absDelta))
		// due to rounding
		if tmp.Cmp(liquidity) > 0 {
			deltaL = tmp.Sub(tmp, liquidity)
		} else {
			deltaL = constants.Zero
		}
	}

	return deltaL
}

// estimateIncrementalLiquidity estimates deltaL, the swap fee to be collected based on amount specified
// for the final swap step to be performed,
// where the next (temporary) tick will not be crossed
func estimateIncrementalLiquidity(
	absAmount, liquidity, sqrtRatioCurrentX96 *int256.Int, feeInUnits constants.FeeAmount, exactIn, isToken0 bool,
) (deltaL *int256.Int) {
	// this is when we didn't reach the target (last step before loop end), then we have to recalculate the target_X96, deltaL ...
	fee := int256.NewInt(int64(feeInUnits))

	if exactIn {
		if isToken0 {
			// deltaL = feeInFeeUnits * absDelta * currentSqrtP / 2
			deltaL = MulDiv(sqrtRatioCurrentX96, int256.New().Mul(absAmount, fee), int256.New().Lsh(TwoFeeUnits, 96))
		} else {
			// deltaL = feeInFeeUnits * absDelta * / (currentSqrtP * 2)
			// Because nextSqrtP = (liquidity + absDelta / currentSqrtP) * currentSqrtP / (liquidity + deltaL)
			// so we round down deltaL, to round up nextSqrtP
			deltaL = MulDivRoundingDown(
				constants.Q96, int256.New().Mul(absAmount, fee), int256.New().Mul(TwoFeeUnits, sqrtRatioCurrentX96),
			)
		}
	} else {
		// obtain the smaller root of the quadratic equation
		// ax^2 - 2bx + c = 0 such that b > 0, and x denotes deltaL
		tmp := int256.New().Sub(FeeUnits, fee)
		a := fee
		b := int256.New().Mul(tmp, liquidity)
		c := int256.New().Mul(tmp.Mul(fee, liquidity), absAmount)

		if isToken0 {
			// a = feeInFeeUnits
			// b = (FEE_UNITS - feeInFeeUnits) * liquidity - FEE_UNITS * absDelta * currentSqrtP
			// c = feeInFeeUnits * liquidity * absDelta * currentSqrtP
			b = tmp.Sub(b, MulDiv(tmp.Mul(FeeUnits, absAmount), sqrtRatioCurrentX96, constants.Q96))
			c = MulDiv(c, sqrtRatioCurrentX96, constants.Q96)
		} else {
			// a = feeInFeeUnits
			// b = (FEE_UNITS - feeInFeeUnits) * liquidity - FEE_UNITS * absDelta / currentSqrtP
			// c = liquidity * feeInFeeUnits * absDelta / currentSqrtP
			b = tmp.Sub(b, MulDiv(tmp.Mul(FeeUnits, absAmount), constants.Q96, sqrtRatioCurrentX96))
			c = MulDiv(c, constants.Q96, sqrtRatioCurrentX96)
		}

		deltaL = GetSmallerRootOfQuadEqn(a, b, c)
	}

	return deltaL
}

// calcFinalPrice calculates returned output | input tokens in exchange for specified amount
// round down when calculating returned output (isExactInput) so we avoid sending too much
// round up when calculating returned input (!isExactInput) so we get desired output amount
func calcFinalPrice(
	absAmount, liquidity, deltaL, sqrtRatioCurrentX96 *int256.Int, exactIn, isToken0 bool,
) (returnAmount *int256.Int) {
	if isToken0 {
		tmp := MulDiv(absAmount, sqrtRatioCurrentX96, constants.Q96)

		if exactIn {
			// minimise actual output (<0, make less negative) so we avoid sending too much
			// returnedAmount = deltaL * nextSqrtP - liquidity * (currentSqrtP - nextSqrtP)
			returnAmount = MulDivRoundingUp(
				int256.New().Add(liquidity, deltaL), sqrtRatioCurrentX96, tmp.Add(liquidity, tmp),
			)
		} else {
			// maximise actual input (>0) so we get desired output amount
			// returnedAmount = deltaL * nextSqrtP + liquidity * (nextSqrtP - currentSqrtP)
			returnAmount = MulDiv(
				int256.New().Add(liquidity, deltaL), sqrtRatioCurrentX96, tmp.Sub(liquidity, tmp),
			)
		}
	} else {
		// returnedAmount = (liquidity + deltaL)/nextSqrtP - (liquidity)/currentSqrtP
		// if exactInput, minimise actual output (<0, make less negative) so we avoid sending too much
		// if exactOutput, maximise actual input (>0) so we get desired output amount
		tmp := MulDiv(absAmount, constants.Q96, sqrtRatioCurrentX96)

		if exactIn {
			returnAmount = MulDiv(
				tmp.Add(liquidity, tmp), sqrtRatioCurrentX96, int256.New().Add(liquidity, deltaL),
			)
		} else {
			returnAmount = MulDivRoundingUp(
				tmp.Sub(liquidity, tmp), sqrtRatioCurrentX96, int256.New().Add(liquidity, deltaL),
			)
		}
	}

	if exactIn && returnAmount.Cmp(constants.One) == 0 {
		returnAmount = constants.Zero
	}

	return returnAmount
}
