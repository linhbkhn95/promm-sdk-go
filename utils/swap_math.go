package utils

import (
	"math/big"

	"github.com/KyberNetwork/promm-sdk-go/constants"
)

var FeeUnits = new(big.Int).Exp(big.NewInt(10), big.NewInt(5), nil)
var TwoFeeUnits = new(big.Int).Mul(FeeUnits, big.NewInt(2))

// ComputeSwapStep computes the actual swap input / output amounts to be deducted or added,
// the swap fee to be collected and the resulting sqrtP
func ComputeSwapStep(
	sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, amountRemaining *big.Int, feeInUnits constants.FeeAmount,
	exactIn, isToken0 bool,
) (sqrtRatioNextX96, amountIn, amountOut, deltaL *big.Int, err error) {
	// in the event currentSqrtP == targetSqrtP because of tick movements, return
	// e.g. swapped up tick where specified price limit is on an initialised tick
	// then swapping down tick will cause next tick to be the same as the current tick
	if sqrtRatioCurrentX96.Cmp(sqrtRatioTargetX96) == 0 {
		return sqrtRatioCurrentX96, constants.Zero, constants.Zero, constants.Zero, nil
	}

	sqrtRatioNextX96 = big.NewInt(0)

	usedAmount := calcReachAmount(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, feeInUnits, exactIn, isToken0)

	if exactIn && usedAmount.Cmp(amountRemaining) >= 0 || (!exactIn && usedAmount.Cmp(amountRemaining) <= 0) {
		usedAmount = amountRemaining
	} else {
		sqrtRatioNextX96 = sqrtRatioTargetX96
	}

	// ??? not use
	amountIn = usedAmount

	var absUsedAmount *big.Int

	if usedAmount.Cmp(constants.Zero) >= 0 {
		absUsedAmount = usedAmount
	} else {
		absUsedAmount = new(big.Int).Mul(usedAmount, constants.NegativeOne)
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
	sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity *big.Int, feeInUnits constants.FeeAmount,
	exactIn, isToken0 bool,
) (reachAmount *big.Int) {
	var absPriceDiff *big.Int

	if sqrtRatioCurrentX96.Cmp(sqrtRatioTargetX96) >= 0 {
		absPriceDiff = new(big.Int).Sub(sqrtRatioCurrentX96, sqrtRatioTargetX96)
	} else {
		absPriceDiff = new(big.Int).Sub(sqrtRatioTargetX96, sqrtRatioCurrentX96)
	}

	if exactIn {
		// we round down so that we avoid taking giving away too much for the specified input
		// i.e. require less input qty to move ticks
		if isToken0 {
			// exactInput + swap 0 -> 1
			// numerator = 2 * liquidity * absPriceDiff
			// denominator = currentSqrtP * (2 * targetSqrtP - currentSqrtP * feeInFeeUnits / FEE_UNITS)
			// overflow should not happen because the absPriceDiff is capped to ~5%
			temp := new(big.Int).Mul(TwoFeeUnits, sqrtRatioTargetX96)

			denominator := big.NewInt(int64(feeInUnits))
			denominator.Mul(denominator, sqrtRatioCurrentX96).Sub(temp, denominator)
			numerator := temp.Mul(TwoFeeUnits, absPriceDiff).Mul(temp, liquidity).Div(temp, denominator)

			reachAmount = numerator.Mul(numerator, constants.Q96).Div(numerator, sqrtRatioCurrentX96)
		} else {
			// exactInput + swap 1 -> 0
			// numerator: liquidity * absPriceDiff * (TWO_FEE_UNITS * targetSqrtP - feeInFeeUnits * (targetSqrtP + currentSqrtP))
			// denominator: (TWO_FEE_UNITS * targetSqrtP - feeInFeeUnits * currentSqrtP)
			// overflow should not happen because the absPriceDiff is capped to ~5%
			temp := new(big.Int).Mul(TwoFeeUnits, sqrtRatioCurrentX96)

			denominator := big.NewInt(int64(feeInUnits))
			denominator.Mul(denominator, sqrtRatioTargetX96).Sub(temp, denominator)
			numerator := temp.Mul(TwoFeeUnits, absPriceDiff).Mul(temp, liquidity).Div(temp, denominator)

			reachAmount = numerator.Mul(numerator, sqrtRatioCurrentX96).Div(numerator, constants.Q96)
		}
	} else {
		// we will perform negation as the last step
		// we round down so that we require less output qty to move ticks
		if isToken0 {
			// exactOut + swap 0 -> 1
			// numerator: (liquidity)(absPriceDiff)(2 * currentSqrtP - deltaL * (currentSqrtP + targetSqrtP))
			// denominator: (currentSqrtP * targetSqrtP) * (2 * currentSqrtP - deltaL * targetSqrtP)
			// overflow should not happen because the absPriceDiff is capped to ~5%

			temp := new(big.Int).Mul(TwoFeeUnits, sqrtRatioCurrentX96)

			denominator := big.NewInt(int64(feeInUnits))
			denominator.Mul(denominator, sqrtRatioTargetX96).Sub(temp, denominator)

			temp.SetInt64(int64(feeInUnits)).Mul(temp, sqrtRatioCurrentX96)
			numerator := new(big.Int).Sub(denominator, temp)
			numerator.Mul(numerator, temp.Lsh(liquidity, 96)).Div(numerator, denominator)

			reachAmount = numerator.Mul(numerator, absPriceDiff).Div(numerator, sqrtRatioCurrentX96).Div(numerator, sqrtRatioTargetX96).Mul(numerator, constants.NegativeOne)
		} else {
			// exactOut + swap 1 -> 0
			// numerator: liquidity * absPriceDiff * (TWO_FEE_UNITS * targetSqrtP - feeInFeeUnits * (targetSqrtP + currentSqrtP))
			// denominator: (TWO_FEE_UNITS * targetSqrtP - feeInFeeUnits * currentSqrtP)
			// overflow should not happen because the absPriceDiff is capped to ~5%
			temp := new(big.Int).Mul(TwoFeeUnits, sqrtRatioTargetX96)

			denominator := big.NewInt(int64(feeInUnits))
			denominator.Mul(denominator, sqrtRatioCurrentX96).Sub(temp, denominator)

			temp.SetInt64(int64(feeInUnits)).Mul(temp, sqrtRatioTargetX96)
			numerator := new(big.Int).Sub(denominator, temp)
			numerator.Mul(numerator, liquidity).Div(numerator, denominator)

			reachAmount = numerator.Mul(numerator, absPriceDiff).Div(numerator, constants.Q96).Mul(numerator, constants.NegativeOne)
		}
	}

	return reachAmount
}

// calcReturnedAmount calculates returned output | input tokens in exchange for specified amount
// round down when calculating returned output (isExactInput) so we avoid sending too much
// round up when calculating returned input (!isExactInput) so we get desired output amount
func calcReturnedAmount(
	sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, deltaL *big.Int, exactIn, isToken0 bool,
) (returnedAmount *big.Int) {
	if isToken0 {
		if exactIn {
			// minimise actual output (<0, make less negative) so we avoid sending too much
			// returnedAmount = deltaL * nextSqrtP - liquidity * (currentSqrtP - nextSqrtP)
			returnedAmount = new(big.Int)
			returnedAmount.Sub(sqrtRatioCurrentX96, sqrtRatioTargetX96).Mul(returnedAmount, liquidity).Div(returnedAmount, constants.Q96).Mul(returnedAmount, constants.NegativeOne).Add(returnedAmount, MulDivRoundingUp(deltaL, sqrtRatioTargetX96, constants.Q96))
		} else {
			// maximise actual input (>0) so we get desired output amount
			// returnedAmount = deltaL * nextSqrtP + liquidity * (nextSqrtP - currentSqrtP)
			returnedAmount = new(big.Int)
			returnedAmount.Sub(sqrtRatioTargetX96, sqrtRatioCurrentX96)
			returnedAmount.Add(MulDivRoundingUp(deltaL, sqrtRatioTargetX96, constants.Q96), MulDivRoundingUp(liquidity, returnedAmount, constants.Q96))
		}
	} else {
		// returnedAmount = (liquidity + deltaL)/nextSqrtP - (liquidity)/currentSqrtP
		// if exactInput, minimise actual output (<0, make less negative) so we avoid sending too much
		// if exactOutput, maximise actual input (>0) so we get desired output amount
		temp := new(big.Int).Add(liquidity, deltaL)
		temp2 := MulDivRoundingUp(liquidity, constants.Q96, sqrtRatioCurrentX96)
		returnedAmount = temp.Add(
			MulDivRoundingUp(temp, constants.Q96, sqrtRatioTargetX96),
			temp2.Mul(temp2, constants.NegativeOne),
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
	sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, absAmount *big.Int, exactIn, isToken0 bool,
) (deltaL *big.Int) {

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
		tmp.Mul(tmp, sqrtRatioTargetX96).Div(tmp, constants.Q96)

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
		tmp.Mul(tmp, constants.Q96).Div(tmp, sqrtRatioTargetX96)

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
	absAmount, liquidity, sqrtRatioCurrentX96 *big.Int, feeInUnits constants.FeeAmount, exactIn, isToken0 bool,
) (deltaL *big.Int) {
	// this is when we didn't reach the target (last step before loop end), then we have to recalculate the target_X96, deltaL ...
	fee := big.NewInt(int64(feeInUnits))

	if exactIn {
		if isToken0 {
			// deltaL = feeInFeeUnits * absDelta * currentSqrtP / 2
			deltaL = MulDiv(sqrtRatioCurrentX96, new(big.Int).Mul(absAmount, fee), new(big.Int).Lsh(TwoFeeUnits, 96))
		} else {
			// deltaL = feeInFeeUnits * absDelta * / (currentSqrtP * 2)
			// Because nextSqrtP = (liquidity + absDelta / currentSqrtP) * currentSqrtP / (liquidity + deltaL)
			// so we round down deltaL, to round up nextSqrtP
			deltaL = MulDivRoundingDown(
				constants.Q96, new(big.Int).Mul(absAmount, fee), new(big.Int).Mul(TwoFeeUnits, sqrtRatioCurrentX96),
			)
		}
	} else {
		// obtain the smaller root of the quadratic equation
		// ax^2 - 2bx + c = 0 such that b > 0, and x denotes deltaL
		tmp := new(big.Int).Sub(FeeUnits, fee)
		a := fee
		b := new(big.Int).Mul(tmp, liquidity)
		c := new(big.Int).Mul(tmp.Mul(fee, liquidity), absAmount)

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
	absAmount, liquidity, deltaL, sqrtRatioCurrentX96 *big.Int, exactIn, isToken0 bool,
) (returnAmount *big.Int) {
	if isToken0 {
		tmp := MulDiv(absAmount, sqrtRatioCurrentX96, constants.Q96)

		if exactIn {
			// minimise actual output (<0, make less negative) so we avoid sending too much
			// returnedAmount = deltaL * nextSqrtP - liquidity * (currentSqrtP - nextSqrtP)
			returnAmount = MulDivRoundingUp(
				new(big.Int).Add(liquidity, deltaL), sqrtRatioCurrentX96, tmp.Add(liquidity, tmp),
			)
		} else {
			// maximise actual input (>0) so we get desired output amount
			// returnedAmount = deltaL * nextSqrtP + liquidity * (nextSqrtP - currentSqrtP)
			returnAmount = MulDiv(
				new(big.Int).Add(liquidity, deltaL), sqrtRatioCurrentX96, tmp.Sub(liquidity, tmp),
			)
		}
	} else {
		// returnedAmount = (liquidity + deltaL)/nextSqrtP - (liquidity)/currentSqrtP
		// if exactInput, minimise actual output (<0, make less negative) so we avoid sending too much
		// if exactOutput, maximise actual input (>0) so we get desired output amount
		tmp := MulDiv(absAmount, constants.Q96, sqrtRatioCurrentX96)

		if exactIn {
			returnAmount = MulDiv(
				tmp.Add(liquidity, tmp), sqrtRatioCurrentX96, new(big.Int).Add(liquidity, deltaL),
			)
		} else {
			returnAmount = MulDivRoundingUp(
				tmp.Sub(liquidity, tmp), sqrtRatioCurrentX96, new(big.Int).Add(liquidity, deltaL),
			)
		}
	}

	if exactIn && returnAmount.Cmp(constants.One) == 0 {
		returnAmount = constants.Zero
	}

	return returnAmount
}
