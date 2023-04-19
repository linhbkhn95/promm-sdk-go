module github.com/KyberNetwork/promm-sdk-go

go 1.17

replace github.com/daoleno/uniswap-sdk-core v0.1.5 => github.com/KyberNetwork/uniswap-sdk-core v0.1.5

replace github.com/linhbkhn95/int256 v0.0.0-20230418034132-fb7f283e6e2a => /Users/kevin/Desktop/Code/me/int256

require (
	github.com/daoleno/uniswap-sdk-core v0.1.5
	github.com/ethereum/go-ethereum v1.10.20
	github.com/shopspring/decimal v1.3.1
	github.com/stretchr/testify v1.8.0
)

require github.com/linhbkhn95/int256 v0.0.0-20230418034132-fb7f283e6e2a

require (
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/holiman/uint256 v1.2.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/sys v0.0.0-20220702020025-31831981b65f // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
