package client

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"log"
	"math/big"
	"sort"
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

func (c *Client) getDelegations(collator string) ([]DelegatorState, error) {
	account, _ := types.HexDecodeString(collator)
	cd := make([]DelegatorState, 0)
	var delegations struct {
		Delegations []candidateDelegationUnmarshal
	}
	err := c.GetStorageRawAt(
		"ParachainStaking",
		"TopDelegations",
		"Delegations<Balance>",
		c.SnapBlock.Hash,
		&delegations,
		account,
	)
	if err != nil {
		log.Printf("Cannot load delegations %v\n", err)
		return nil, err
	}
	// Fetch delegations
	requests := make([]delegationScheduledRequestsUnmarshal, 0)
	err = c.GetStorageRawAt(
		"ParachainStaking",
		"DelegationScheduledRequests",
		"Vec<DelegationScheduledRequests<DelegatorState<Balance>>>",
		c.SnapBlock.Hash,
		&requests,
		account,
	)
	if err != nil {
		log.Printf("Unable to decode delegator scheduled requests for %v\n", collator)
		return nil, err
	}
	// Get state
	for _, delegation := range delegations.Delegations {
		cd = append(cd, c.getDelegatorState(requests, delegation.Owner, delegation.Amount))
	}
	// Sort
	sort.Slice(cd[:], func(i, j int) bool {
		return cd[i].Amount.Balance.Cmp(cd[j].Amount.Balance) == 1
	})
	// Done
	return cd, nil
}

func (c *Client) getDelegatorState(
	requests []delegationScheduledRequestsUnmarshal,
	address string,
	total TokenAmount,
) DelegatorState {
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
	return r
}
