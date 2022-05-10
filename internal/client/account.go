package client

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/zooper-corp/mooncli/internal/async"
)

type AccountInfo struct {
	Address  string          `json:"address"`
	Balance  AccountBalance  `json:"balance"`
	Identity AccountIdentity `json:"identity,omitempty"`
}

func (c *Client) FetchAccountInfo(address string) (AccountInfo, error) {
	account, err := types.HexDecodeString(address)
	if err != nil {
		return AccountInfo{}, err
	}
	// Fetch balance
	balanceChannel := make(chan async.Result[AccountBalance])
	go func() {
		balanceChannel <- async.ResultFrom(c.accountBalanceFromAccount(account))
	}()
	// Fetch identity
	identityChannel := make(chan async.Result[AccountIdentity])
	go func() {
		identityChannel <- async.ResultFrom(c.accountIdentityFromAccount(account))
	}()
	// Collect
	result := AccountInfo{}
	balance := <-balanceChannel
	if balance.IsErr() {
		return result, balance.Err
	} else {
		result.Balance = balance.Value
	}
	identity := <-identityChannel
	// Ignore error on identity as its optional
	if !identity.IsErr() {
		result.Identity = identity.Value
	}
	// All good
	return result, nil
}
