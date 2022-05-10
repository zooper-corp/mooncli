package client

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type stakingRoundInfo struct {
	Current types.U32
	First   types.U32
	Length  types.U32
}

func fetchStakingRoundInfo(c *Client, hash types.Hash) (stakingRoundInfo, error) {
	key, err := types.CreateStorageKey(c.metadata, "ParachainStaking", "Round", nil, nil)
	if err != nil {
		return stakingRoundInfo{}, err
	}
	var roundInfo stakingRoundInfo
	ok, err := c.api.RPC.State.GetStorage(key, &roundInfo, hash)
	if err != nil || !ok {
		return stakingRoundInfo{}, err
	}
	return roundInfo, nil
}
