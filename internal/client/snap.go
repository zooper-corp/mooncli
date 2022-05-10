package client

import (
	"fmt"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"math"
)

type Snap struct {
	Block   SnapBlock
	Round   SnapRound
	Staking SnapStaking
}

type SnapBlock struct {
	Number       uint64     `json:"current"`
	Hash         types.Hash `json:"hash"`
	TsMillis     uint64     `json:"ts"`
	DurationSecs float64    `json:"duration"`
}

type SnapRound struct {
	Number      uint32 `json:"number"`
	Length      uint32 `json:"length"`
	Start       uint32 `json:"start"`
	RevokeDelay uint32 `json:"revoke_delay"`
}

type SnapStaking struct {
	Selected uint32 `json:"selected"`
	Total    uint32 `json:"total"`
}

func fetchBlockTs(c *Client, blockHash types.Hash) (uint64, error) {
	key, err := types.CreateStorageKey(c.metadata, "Timestamp", "Now", nil, nil)
	if err != nil {
		return 0, err
	}
	var blockTs types.U64
	ok, err := c.api.RPC.State.GetStorage(key, &blockTs, blockHash)
	if err != nil || !ok {
		return 0, err
	}
	return uint64(blockTs), nil
}

func fetchSnapBlock(c *Client, targetBlock int64, targetRound uint32) (Snap, error) {
	// Get head block
	blockHash, err := c.api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		return Snap{}, err
	}
	// Get block number
	blockNumber, err := c.GetBlockNumber(blockHash)
	if err != nil {
		return Snap{}, err
	}
	// Go to target
	if targetRound != 0 || targetBlock != 0 {
		if targetRound == 0 {
			if uint64(targetBlock) > blockNumber {
				return Snap{}, fmt.Errorf("invalid block %v > %v", targetBlock, blockNumber)
			}
			blockNumber = uint64(targetBlock)
			blockHash, err = c.api.RPC.Chain.GetBlockHash(blockNumber)
			if err != nil {
				return Snap{}, err
			}
		} else {
			currentRoundInfo, err := fetchStakingRoundInfo(c, blockHash)
			if err != nil {
				return Snap{}, err
			}
			currentRound := uint32(currentRoundInfo.Current)
			if targetRound <= currentRound {
				targetRoundBlock := blockNumber - (uint64(currentRoundInfo.Length) * uint64(currentRound-targetRound))
				targetRoundHash, err := c.api.RPC.Chain.GetBlockHash(targetRoundBlock)
				if err != nil {
					return Snap{}, err
				}
				targetRoundInfo, err := fetchStakingRoundInfo(c, targetRoundHash)
				if err != nil {
					return Snap{}, err
				}
				// Set relative block and number, the reset hash
				blockNumber = uint64(targetRoundInfo.First)
				if targetBlock != 0 {
					blockNumber = uint64(int64(blockNumber) + targetBlock)
				}
				targetHash, err := c.api.RPC.Chain.GetBlockHash(blockNumber)
				if err != nil {
					return Snap{}, err
				}
				blockHash = targetHash
			}
		}
	}
	// Fetch round info at target block
	roundInfo, err := fetchStakingRoundInfo(c, blockHash)
	if err != nil {
		return Snap{}, err
	}
	// Fetch block TS
	blockTs, err := fetchBlockTs(c, blockHash)
	if err != nil {
		return Snap{}, err
	}
	// Fetch average
	blockDelta := math.Min(float64(blockNumber)-1, float64(10000))
	hashPast, err := c.api.RPC.Chain.GetBlockHash(blockNumber - uint64(blockDelta))
	if err != nil {
		return Snap{}, err
	}
	blockPastTs, err := fetchBlockTs(c, hashPast)
	if err != nil {
		return Snap{}, err
	}
	// Fetch staking pool data
	pool, err := c.FetchSortedCandidatePool(blockHash)
	if err != nil {
		return Snap{}, err
	}
	selected, err := c.FetchSelectedCandidates(blockHash)
	if err != nil {
		return Snap{}, err
	}
	// Bond less delay
	bondLessDelay, err := c.GetCandidateBondLessDelay()
	if err != nil {
		return Snap{}, err
	}
	// At target already
	return Snap{
		Block: SnapBlock{
			Number:       blockNumber,
			DurationSecs: float64(blockTs-blockPastTs) / blockDelta / 1000.0,
			Hash:         blockHash,
			TsMillis:     blockTs,
		},
		Round: SnapRound{
			Number:      uint32(roundInfo.Current),
			Length:      uint32(roundInfo.Length),
			Start:       uint32(roundInfo.First),
			RevokeDelay: bondLessDelay,
		},
		Staking: SnapStaking{
			Total:    uint32(len(pool)),
			Selected: uint32(len(selected)),
		},
	}, nil
}
