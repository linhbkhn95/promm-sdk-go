package utils

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/linhbkhn95/int256"

	"github.com/KyberNetwork/promm-sdk-go/constants"
)

const (
	MinTick = -887272  // The minimum tick that can be used on any pool.
	MaxTick = -MinTick // The maximum tick that can be used on any pool.
)

var (
	Q32             = int256.NewInt(1 << 32)
	MinSqrtRatio    = int256.NewInt(4295128739)                                                   // The sqrt ratio corresponding to the minimum tick that could be used on any pool.
	MaxSqrtRatio, _ = int256.New().SetString("1461446703485210103287273052203988822378723970342") // The sqrt ratio corresponding to the maximum tick that could be used on any pool.
)

var (
	ErrInvalidTick      = errors.New("invalid tick")
	ErrInvalidSqrtRatio = errors.New("invalid sqrt ratio")
)

func mulShift(val *int256.Int, mulBy *int256.Int) *int256.Int {
	temp := int256.New().Mul(val, mulBy)
	return temp.Rsh(temp, 128)
}

var (
	sqrtConst1BigInt, _  = new(big.Int).SetString("fffcb933bd6fad37aa2d162d1a594001", 16)
	sqrtConst2BigInt, _  = new(big.Int).SetString("100000000000000000000000000000000", 16)
	sqrtConst3BigInt, _  = new(big.Int).SetString("fff97272373d413259a46990580e213a", 16)
	sqrtConst4BigInt, _  = new(big.Int).SetString("fff2e50f5f656932ef12357cf3c7fdcc", 16)
	sqrtConst5BigInt, _  = new(big.Int).SetString("ffe5caca7e10e4e61c3624eaa0941cd0", 16)
	sqrtConst6BigInt, _  = new(big.Int).SetString("ffcb9843d60f6159c9db58835c926644", 16)
	sqrtConst7BigInt, _  = new(big.Int).SetString("ff973b41fa98c081472e6896dfb254c0", 16)
	sqrtConst8BigInt, _  = new(big.Int).SetString("ff2ea16466c96a3843ec78b326b52861", 16)
	sqrtConst9BigInt, _  = new(big.Int).SetString("fe5dee046a99a2a811c461f1969c3053", 16)
	sqrtConst10BigInt, _ = new(big.Int).SetString("fcbe86c7900a88aedcffc83b479aa3a4", 16)
	sqrtConst11BigInt, _ = new(big.Int).SetString("f987a7253ac413176f2b074cf7815e54", 16)
	sqrtConst12BigInt, _ = new(big.Int).SetString("f3392b0822b70005940c7a398e4b70f3", 16)
	sqrtConst13BigInt, _ = new(big.Int).SetString("e7159475a2c29b7443b29c7fa6e889d9", 16)
	sqrtConst14BigInt, _ = new(big.Int).SetString("d097f3bdfd2022b8845ad8f792aa5825", 16)
	sqrtConst15BigInt, _ = new(big.Int).SetString("a9f746462d870fdf8a65dc1f90e061e5", 16)
	sqrtConst16BigInt, _ = new(big.Int).SetString("70d869a156d2a1b890bb3df62baf32f7", 16)
	sqrtConst17BigInt, _ = new(big.Int).SetString("31be135f97d08fd981231505542fcfa6", 16)
	sqrtConst18BigInt, _ = new(big.Int).SetString("9aa508b5b7a84e1c677de54f3e99bc9", 16)
	sqrtConst19BigInt, _ = new(big.Int).SetString("5d6af8dedb81196699c329225ee604", 16)
	sqrtConst20BigInt, _ = new(big.Int).SetString("2216e584f5fa1ea926041bedfe98", 16)
	sqrtConst21BigInt, _ = new(big.Int).SetString("48a170391f7dc42444e8fa2", 16)

	sqrtConst1  = int256.MustFromBig(sqrtConst1BigInt)
	sqrtConst2  = int256.MustFromBig(sqrtConst2BigInt)
	sqrtConst3  = int256.MustFromBig(sqrtConst3BigInt)
	sqrtConst4  = int256.MustFromBig(sqrtConst4BigInt)
	sqrtConst5  = int256.MustFromBig(sqrtConst5BigInt)
	sqrtConst6  = int256.MustFromBig(sqrtConst6BigInt)
	sqrtConst7  = int256.MustFromBig(sqrtConst7BigInt)
	sqrtConst8  = int256.MustFromBig(sqrtConst8BigInt)
	sqrtConst9  = int256.MustFromBig(sqrtConst9BigInt)
	sqrtConst10 = int256.MustFromBig(sqrtConst10BigInt)
	sqrtConst11 = int256.MustFromBig(sqrtConst11BigInt)
	sqrtConst12 = int256.MustFromBig(sqrtConst12BigInt)
	sqrtConst13 = int256.MustFromBig(sqrtConst13BigInt)
	sqrtConst14 = int256.MustFromBig(sqrtConst14BigInt)
	sqrtConst15 = int256.MustFromBig(sqrtConst15BigInt)
	sqrtConst16 = int256.MustFromBig(sqrtConst16BigInt)
	sqrtConst17 = int256.MustFromBig(sqrtConst17BigInt)
	sqrtConst18 = int256.MustFromBig(sqrtConst18BigInt)
	sqrtConst19 = int256.MustFromBig(sqrtConst19BigInt)
	sqrtConst20 = int256.MustFromBig(sqrtConst20BigInt)
	sqrtConst21 = int256.MustFromBig(sqrtConst21BigInt)
	maxInt256   = int256.MustFromBig(entities.MaxUint256)
)

/**
 * Returns the sqrt ratio as a Q64.96 for the given tick. The sqrt ratio is computed as sqrt(1.0001)^tick
 * @param tick the tick for which to compute the sqrt ratio
 */
func GetSqrtRatioAtTick(tick int) (*int256.Int, error) {
	if tick < MinTick || tick > MaxTick {
		return nil, ErrInvalidTick
	}
	absTick := tick
	if tick < 0 {
		absTick = -tick
	}
	var ratio *int256.Int
	if absTick&0x1 != 0 {
		ratio = sqrtConst1
	} else {
		ratio = sqrtConst2
	}
	if (absTick & 0x2) != 0 {
		ratio = mulShift(ratio, sqrtConst3)
	}
	if (absTick & 0x4) != 0 {
		ratio = mulShift(ratio, sqrtConst4)
	}
	if (absTick & 0x8) != 0 {
		ratio = mulShift(ratio, sqrtConst5)
	}
	if (absTick & 0x10) != 0 {
		ratio = mulShift(ratio, sqrtConst6)
	}
	if (absTick & 0x20) != 0 {
		ratio = mulShift(ratio, sqrtConst7)
	}
	if (absTick & 0x40) != 0 {
		ratio = mulShift(ratio, sqrtConst8)
	}
	if (absTick & 0x80) != 0 {
		ratio = mulShift(ratio, sqrtConst9)
	}
	if (absTick & 0x100) != 0 {
		ratio = mulShift(ratio, sqrtConst10)
	}
	if (absTick & 0x200) != 0 {
		ratio = mulShift(ratio, sqrtConst11)
	}
	if (absTick & 0x400) != 0 {
		ratio = mulShift(ratio, sqrtConst12)
	}
	if (absTick & 0x800) != 0 {
		ratio = mulShift(ratio, sqrtConst13)
	}
	if (absTick & 0x1000) != 0 {
		ratio = mulShift(ratio, sqrtConst14)
	}
	if (absTick & 0x2000) != 0 {
		ratio = mulShift(ratio, sqrtConst15)
	}
	if (absTick & 0x4000) != 0 {
		ratio = mulShift(ratio, sqrtConst16)
	}
	if (absTick & 0x8000) != 0 {
		ratio = mulShift(ratio, sqrtConst17)
	}
	if (absTick & 0x10000) != 0 {
		ratio = mulShift(ratio, sqrtConst18)
	}
	if (absTick & 0x20000) != 0 {
		ratio = mulShift(ratio, sqrtConst19)
	}
	if (absTick & 0x40000) != 0 {
		ratio = mulShift(ratio, sqrtConst20)
	}
	if (absTick & 0x80000) != 0 {
		ratio = mulShift(ratio, sqrtConst21)
	}
	if tick > 0 {
		ratio = int256.New().Div(maxInt256, ratio)
	}

	// back to Q96
	if int256.New().Rem(ratio, Q32).Cmp(constants.Zero) > 0 {
		return int256.New().Add((int256.New().Div(ratio, Q32)), constants.One), nil
	} else {
		return int256.New().Div(ratio, Q32), nil
	}
}

var (
	magicSqrt10001, _ = int256.New().SetString("255738958999603826347141")
	magicTickLow, _   = int256.New().SetString("3402992956809132418596140100660247210")
	magicTickHigh, _  = int256.New().SetString("291339464771989622907027621153398088495")
)

/**
 * Returns the tick corresponding to a given sqrt ratio, s.t. #getSqrtRatioAtTick(tick) <= sqrtRatioX96
 * and #getSqrtRatioAtTick(tick + 1) > sqrtRatioX96
 * @param sqrtRatioX96 the sqrt ratio as a Q64.96 for which to compute the tick
 */
func GetTickAtSqrtRatio(sqrtRatioX96 *int256.Int) (int, error) {
	if sqrtRatioX96.Cmp(MinSqrtRatio) < 0 || sqrtRatioX96.Cmp(MaxSqrtRatio) >= 0 {
		return 0, ErrInvalidSqrtRatio
	}
	sqrtRatioX128 := int256.New().Lsh(sqrtRatioX96, 32)
	msb, err := MostSignificantBit(sqrtRatioX128)
	if err != nil {
		return 0, err
	}
	var r *int256.Int
	if int256.NewInt(msb).Cmp(constants.BigInt128) >= 0 {
		r = sqrtRatioX128.Rsh(sqrtRatioX128, uint(msb-127))
	} else {
		r = sqrtRatioX128.Lsh(sqrtRatioX128, uint(127-msb))
	}
	msbBigInt := int256.NewInt(msb)
	msbBigInt.Sub(msbBigInt, constants.BigInt128)
	log2 := msbBigInt.Lsh(msbBigInt, 64)

	for i := 0; i < 14; i++ {
		r.Rsh(r.Mul(r, r), 127)
		f := int256.New().Rsh(r, 128)
		log2 = int256.New().Or(log2, int256.New().Lsh(f, uint(63-i)))
		r.Rsh(r, uint(f.Int64()))
	}

	logSqrt10001 := log2.Mul(log2, magicSqrt10001)
	temp := int256.New().Sub(logSqrt10001, magicTickLow)
	tickLow := temp.Rsh(temp, 128).Int64()
	temp = temp.Add(logSqrt10001, magicTickHigh)
	tickHigh := temp.Rsh(temp, 128).Int64()

	if tickLow == tickHigh {
		return int(tickLow), nil
	}

	sqrtRatio, err := GetSqrtRatioAtTick(int(tickHigh))
	if err != nil {
		return 0, err
	}

	fmt.Println("tickHigh", tickLow, tickHigh, sqrtRatio.String(), sqrtRatioX96.String())
	if sqrtRatio.Cmp(sqrtRatioX96) <= 0 {
		return int(tickHigh), nil
	} else {
		return int(tickLow), nil
	}
}
