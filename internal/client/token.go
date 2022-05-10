package client

import (
	"encoding/json"
	"fmt"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"math"
	"math/big"
	"strings"
)

type tokenInfoUnmarshal struct {
	TokenDecimals uint32
	TokenSymbol   string
}

type TokenInfo struct {
	TokenDecimals uint32 `json:"decimals"`
	TokenSymbol   string `json:"symbol"`
}

type TokenAmount struct {
	int *big.Int
}

func (b *TokenAmount) Cmp(y *TokenAmount) int {
	return b.int.Cmp(y.int)
}

func (b *TokenAmount) AsBigInt() *big.Int {
	r := big.NewInt(0)
	r.Set(b.int)
	return r
}

func (b *TokenAmount) AsBalance(info *TokenInfo) TokenBalance {
	return TokenBalance{
		info:    info,
		Balance: &TokenAmount{b.AsBigInt()},
	}
}

func (b *TokenAmount) MarshalJSON() ([]byte, error) {
	return []byte(b.int.String()), nil
}

func (b *TokenAmount) UnmarshalJSON(p []byte) error {
	if string(p) == "null" {
		return nil
	}
	var z big.Int
	_, ok := z.SetString(strings.Trim(string(p), "\""), 10)
	if !ok {
		return fmt.Errorf("not a valid big integer: %s", p)
	}
	b.int = &z
	return nil
}

func fetchTokenInfo(c *Client) (TokenInfo, error) {
	var t tokenInfoUnmarshal
	err := c.api.Client.Call(&t, "system_properties")
	return TokenInfo(t), err
}

func TokenBalanceU128(c *Client, u128 types.U128) TokenBalance {
	return TokenBalance{
		info:    &c.TokenInfo,
		Balance: &TokenAmount{u128.Int},
	}
}

type TokenBalance struct {
	info    *TokenInfo
	Balance *TokenAmount `json:"balance"`
}

func (tb *TokenBalance) Float64() float64 {
	if tb.info != nil && tb.Balance != nil {
		fb := new(big.Float).SetInt(tb.Balance.int)
		fe := new(big.Float).SetFloat64(math.Pow10(int(tb.info.TokenDecimals) * -1))
		result, _ := new(big.Float).Mul(fb, fe).Float64()
		return result
	} else {
		return 0.0
	}
}

func (tb TokenBalance) MarshalJSON() ([]byte, error) {
	floatValue := tb.Float64()
	if floatValue != 0 {
		return json.Marshal(floatValue)
	} else {
		return json.Marshal(0)
	}
}
