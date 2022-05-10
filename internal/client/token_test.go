package client

import (
	"encoding/json"
	"math"
	"math/big"
	"testing"
)

func TestTokenBalance_Float64(t *testing.T) {
	const b int64 = 2234274829857847390
	balance := TokenBalance{
		info: &TokenInfo{
			TokenDecimals: 18,
			TokenSymbol:   "TEST",
		},
		Balance: &TokenAmount{big.NewInt(b)},
	}
	balanceFloat := balance.Float64()
	balanceFloor := math.Floor(balanceFloat*100) / 100
	if balanceFloor != 2.23 {
		t.Errorf("got %v, wanted > %v", balanceFloor, 2.23)
	}
}

func TestTokenBalance_Float64_Json(t *testing.T) {
	e, _ := big.NewInt(0).SetString("1683563695289220000000000", 10)
	balance := TokenBalance{
		info: &TokenInfo{
			TokenDecimals: 18,
			TokenSymbol:   "TEST",
		},
		Balance: &TokenAmount{e},
	}
	balanceFloat := balance.Float64()
	balanceFloor := int64(math.Floor(balanceFloat))
	if balanceFloor != 1683563 {
		t.Errorf("got %v, wanted > %v", balanceFloor, 1683563)
	}
	j, _ := json.Marshal(balance)
	if string(j) != "1683563.6952892202" {
		t.Errorf("wanted %v, got > %v", "1683563.695", string(j))
	}
}
