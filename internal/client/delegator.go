package client

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"log"
	"math/big"
	"strings"
)

type delegationRequestUnmarshal struct {
	Collator string
	Amount   TokenAmount
	Round    uint32
	Action   ScaleEnum
}

type delegationRequestMapUnmarshal struct {
	Collator string
	Request  delegationRequestUnmarshal
}

type delegationScheduledRequestsUnmarshal struct {
	Delegator string
	Round     uint32
	Action    map[string]TokenAmount
}

type candidateDelegationUnmarshal struct {
	Owner  string
	Amount TokenAmount
}

type delegationRequestsUnmarshal struct {
	Count    uint32
	Requests []delegationRequestMapUnmarshal
	Total    TokenAmount
}

type delegatorStateUnmarshal struct {
	Id          string
	Delegations []candidateDelegationUnmarshal
	Total       TokenAmount
	Requests    delegationRequestsUnmarshal
	Status      ScaleEnum
}

type delegatorStateUnmarshalV1500 struct {
	Id          string
	Delegations []candidateDelegationUnmarshal
	Total       TokenAmount
	LessTotal   TokenAmount
	Status      ScaleEnum
}

type DelegatorState struct {
	Address      string       `json:"address"`
	Amount       TokenBalance `json:"amount"`
	RevokeAmount TokenBalance `json:"revoke_amount,omitempty"`
	RevokeReason string       `json:"revoke_reason,omitempty"`
	RevokeRound  uint32       `json:"revoke_round,omitempty"`
}

func (c *Client) FetchDelegatorState(collator string, address string) (DelegatorState, error) {
	delegatorAccount, _ := types.HexDecodeString(address)
	collatorAccount, _ := types.HexDecodeString(collator)
	var delegator delegatorStateUnmarshalV1500
	err := c.GetStorageRaw(
		"ParachainStaking",
		"DelegatorState",
		"DelegatorState<Balance>",
		&delegator,
		delegatorAccount,
	)
	if err != nil {
		log.Printf("Unable to decode delegator state for %v\n", address)
		return DelegatorState{}, err
	}
	// Fetch collator relative data
	totalDelegated := big.NewInt(0)
	for _, delegation := range delegator.Delegations {
		if strings.EqualFold(delegation.Owner, collator) {
			totalDelegated.Add(totalDelegated, delegation.Amount.AsBigInt())
		}
	}
	// Fetch revokes for collator
	requests := make([]delegationScheduledRequestsUnmarshal, 0)
	err = c.GetStorageRaw(
		"ParachainStaking",
		"DelegationScheduledRequests",
		"Vec<DelegationScheduledRequests<DelegatorState<Balance>>>",
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
			Balance: &TokenAmount{totalDelegated},
		},
		RevokeAmount: revokeAmount.AsBalance(&c.TokenInfo),
		RevokeReason: revokeReason,
		RevokeRound:  revokeRound,
	}
	return r, nil
}
