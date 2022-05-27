package client

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"log"
	"math/big"
	"strings"
)

type delegationScheduledRequestsUnmarshal struct {
	Delegator string
	Round     uint32
	Action    map[string]TokenAmount
}

type candidateDelegationUnmarshal struct {
	Owner  string
	Amount TokenAmount
}

type DelegatorState struct {
	Address      string       `json:"address"`
	Amount       TokenBalance `json:"amount"`
	RevokeAmount TokenBalance `json:"revoke_amount,omitempty"`
	RevokeReason string       `json:"revoke_reason,omitempty"`
	RevokeRound  uint32       `json:"revoke_round,omitempty"`
}

func (c *Client) FetchDelegatorState(collator string, address string, total TokenAmount) (DelegatorState, error) {
	collatorAccount, _ := types.HexDecodeString(collator)
	// Fetch revokes for collator
	requests := make([]delegationScheduledRequestsUnmarshal, 0)
	err := c.GetStorageRawAt(
		"ParachainStaking",
		"DelegationScheduledRequests",
		"Vec<DelegationScheduledRequests<DelegatorState<Balance>>>",
		c.SnapBlock.Hash,
		&requests,
		collatorAccount,
	)
	if err != nil {
		log.Printf("Unable to decode delegator state for %v\n", address)
		return DelegatorState{}, err
	}
	revokeAmount := TokenAmount{big.NewInt(0)}
	revokeReason := ""
	revokeRound := uint32(0)
	for _, request := range requests {
		if strings.EqualFold(request.Delegator, address) {
			for action, amount := range request.Action {
				revokeReason = action
				revokeRound = request.Round
				revokeAmount = amount
				break
			}
		}
	}
	// Ok
	r := DelegatorState{
		Address: address,
		Amount: TokenBalance{
			info:    &c.TokenInfo,
			Balance: &TokenAmount{total.AsBigInt()},
		},
		RevokeAmount: revokeAmount.AsBalance(&c.TokenInfo),
		RevokeReason: revokeReason,
		RevokeRound:  revokeRound,
	}
	return r, nil
}
