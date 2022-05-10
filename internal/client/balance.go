package client

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type accountDataUnmarshal struct {
	Nonce       types.U32
	Consumers   types.U32
	Providers   types.U32
	Sufficients types.U32
	Data        struct {
		Free       types.U128
		Reserved   types.U128
		MiscFrozen types.U128
		FreeFrozen types.U128
	}
}

type AccountBalance struct {
	Free     TokenBalance `json:"free"`
	Reserved TokenBalance `json:"reserved"`
}

func (c *Client) accountBalanceFromAccount(account []byte) (AccountBalance, error) {
	var balance accountDataUnmarshal
	ok, err := c.GetStorage("System", "Account", &balance, account)
	if err != nil || !ok {
		return AccountBalance{}, err
	}
	return AccountBalance{
		Free:     TokenBalanceU128(c, balance.Data.Free),
		Reserved: TokenBalanceU128(c, balance.Data.Reserved),
	}, nil
}
