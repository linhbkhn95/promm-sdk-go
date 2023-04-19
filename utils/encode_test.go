package utils

import (
	"testing"

	"github.com/linhbkhn95/int256"
	"github.com/stretchr/testify/assert"

	"github.com/KyberNetwork/promm-sdk-go/constants"
)

func TestEncodeSqrtRatioX96(t *testing.T) {
	assert.Equal(t, EncodeSqrtRatioX96(int256.NewInt(1), int256.NewInt(1)), constants.Q96, "1/1")

	r0, _ := int256.New().SetString("792281625142643375935439503360")
	assert.Equal(t, EncodeSqrtRatioX96(int256.NewInt(100), int256.NewInt(1)), r0, 10, "100/1")

	r1, _ := int256.New().SetString("7922816251426433759354395033")
	assert.Equal(t, EncodeSqrtRatioX96(int256.NewInt(1), int256.NewInt(100)), r1, 10, "1/100")

	r2, _ := int256.New().SetString("45742400955009932534161870629")
	assert.Equal(t, EncodeSqrtRatioX96(int256.NewInt(111), int256.NewInt(333)), r2, 10, "111/333")

	r3, _ := int256.New().SetString("137227202865029797602485611888")
	assert.Equal(t, EncodeSqrtRatioX96(int256.NewInt(333), int256.NewInt(111)), r3, 10, "333/111")
}
