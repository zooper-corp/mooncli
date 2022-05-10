package client

import (
	"bytes"
	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/zooper-corp/mooncli/config"
	"github.com/zooper-corp/mooncli/internal/async"
	"golang.org/x/exp/slices"
	"log"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"
)

type CandidatePoolEntry struct {
	Owner  string
	Amount TokenAmount
}

type candidateMetadataUnmarshal struct {
	Bond           TokenAmount
	Delegations    uint32
	Counted        TokenAmount
	TopAmount      TokenAmount
	BottomAmount   TokenAmount
	LowestAmount   TokenAmount
	TopCapacity    ScaleEnum
	BottomCapacity ScaleEnum
	Status         ScaleEnum
}

type CollatorInfo struct {
	Address     string                     `json:"address"`
	Selected    bool                       `json:"selected"`
	Rank        uint32                     `json:"rank"`
	Blocks      uint32                     `json:"blocks"`
	Counted     TokenBalance               `json:"counted"`
	MinBond     TokenBalance               `json:"min_bond"`
	Balance     AccountBalance             `json:"balance"`
	Display     string                     `json:"display"`
	History     map[uint32]CollatorHistory `json:"history,omitempty"`
	Delegations []DelegatorState           `json:"-"`
	Revokes     map[uint32]RevokeRound     `json:"revokes,omitempty"`
}

type CollatorHistory struct {
	Rank    uint32       `json:"rank"`
	Blocks  uint32       `json:"blocks"`
	Counted TokenBalance `json:"counted"`
}

type CollatorPool struct {
	SelectedSize uint32         `json:"selected_size"`
	RoundNumber  uint32         `json:"round_number"`
	Collators    []CollatorInfo `json:"collators"`
}

type RevokeRound struct {
	Rank    uint32       `json:"rank"`
	Counted TokenBalance `json:"counted"`
	Amount  TokenBalance `json:"amount"`
}

// FetchSelectedCandidates returns a list of addresses currently selected
func (c *Client) FetchSelectedCandidates(blockHash types.Hash) ([]string, error) {
	// Selected pool
	var selected []string
	err := c.GetStorageRawAt(
		"ParachainStaking",
		"SelectedCandidates",
		"SelectedCandidates",
		blockHash,
		&selected,
	)
	if err != nil {
		return []string{}, err
	}
	return selected, nil
}

// FetchSortedCandidatePool returns the full list of candidates with bonded amount
func (c *Client) FetchSortedCandidatePool(blockHash types.Hash) ([]CandidatePoolEntry, error) {
	var pool []CandidatePoolEntry
	err := c.GetStorageRawAt(
		"ParachainStaking",
		"CandidatePool",
		"CandidatePool",
		blockHash,
		&pool,
	)
	if err != nil {
		return []CandidatePoolEntry{}, err
	}
	// Sort
	sort.Slice(pool[:], func(i, j int) bool {
		return pool[i].Amount.Cmp(&pool[j].Amount) == -1
	})
	return pool, nil
}

// GetCandidateBondLessDelay will fetch current bond delay
func (c *Client) GetCandidateBondLessDelay() (uint32, error) {
	var delay uint32
	err := c.GetConstantValue(
		"ParachainStaking",
		"CandidateBondLessDelay",
		"U32",
		&delay,
	)
	if err != nil {
		return 0, err
	}
	return delay, nil
}

func (ci *CollatorInfo) DisplayName() string {
	name := ci.Address[:6] + "..." + ci.Address[len(ci.Address)-4:]
	if display := ci.Display; display != "" {
		name = display
	}
	return name
}

func (ci *CollatorInfo) RevokeAt(round uint32) *RevokeRound {
	var lastV RevokeRound
	lastK := uint32(0)
	for k, v := range ci.Revokes {
		if k > lastK {
			lastV = v
		}
		if k == round {
			return &v
		}
	}
	// Too late
	return &lastV
}

func (ci *CollatorInfo) AverageBlocks() float32 {
	var r float32
	var c int32
	for _, history := range ci.History {
		if history.Blocks > 0 {
			r = r + float32(history.Blocks)
			c++
		}
	}
	if r > 0 && c > 0 {
		return r / float32(c)
	} else {
		return 0.0
	}
}

func (c *Client) FetchCollatorPool(poolConfig config.CollatorsPoolConfig) (CollatorPool, error) {
	start := time.Now().UnixMilli()
	// Full candidate pool
	pool, err := c.FetchSortedCandidatePool(c.SnapBlock.Hash)
	if err != nil {
		return CollatorPool{}, err
	}
	selected, err := c.FetchSelectedCandidates(c.SnapBlock.Hash)
	if err != nil {
		return CollatorPool{}, err
	}
	// Get details
	ch := make(chan async.Result[CollatorInfo])
	// Get fetch list
	addresses := make([]string, 0)
	for _, poolEntry := range pool {
		if len(poolConfig.Address) == 0 || strings.EqualFold(poolConfig.Address, poolEntry.Owner) {
			addresses = append(addresses, poolEntry.Owner)
		}
	}
	// Query storage
	var wg sync.WaitGroup
	for _, address := range addresses {
		wg.Add(1)
		rank := getAddressRank(pool, address)
		address := address
		selected := slices.Contains(selected, address)
		go func() {
			defer wg.Done()
			ch <- async.ResultFrom(c.FetchCollatorInfo(
				strings.Clone(address),
				selected,
				rank,
				poolConfig,
			))
		}()
	}
	// Wait channel
	go func() {
		wg.Wait()
		close(ch)
		log.Printf("Fetched collator pool in %vsecs\n", float64(time.Now().UnixMilli()-start)/1000.0)
	}()
	// Collect
	result := make([]CollatorInfo, 0)
	for r := range ch {
		if r.Err != nil {
			log.Printf("Unable to fetch collator info %v\n", r.Err)
			return CollatorPool{}, nil
		} else {
			result = append(result, r.Value)
		}
	}
	// Sort
	sort.Slice(result[:], func(i, j int) bool {
		if result[i].Counted.Balance == nil {
			return true
		}
		if result[j].Counted.Balance == nil {
			return false
		}
		return result[i].Counted.Balance.Cmp(result[j].Counted.Balance) == 1
	})
	// Create pool
	collatorPool := CollatorPool{
		SelectedSize: uint32(len(selected)),
		RoundNumber:  c.SnapRound.Number,
		Collators:    result,
	}
	// Compute revokes
	collatorPool.computeRevokes(
		c.SnapRound.Number,
		c.SnapRound.Number+c.SnapRound.RevokeDelay,
	)
	// Done
	return collatorPool, nil
}

func (c *Client) FetchCollatorInfo(
	address string,
	selected bool,
	rank uint32,
	cfg config.CollatorsPoolConfig,
) (CollatorInfo, error) {
	account, _ := types.HexDecodeString(address)
	var candidate candidateMetadataUnmarshal
	err := c.GetStorageRaw(
		"ParachainStaking",
		"CandidateInfo",
		"CandidateMetadata<Balance>",
		&candidate,
		account,
	)
	if err != nil {
		return CollatorInfo{}, err
	}
	// Get identity
	info, err := c.FetchAccountInfo(address)
	if err != nil {
		return CollatorInfo{}, err
	}
	// Get historyRounds
	history, err := c.FetchCollatorHistory(address, cfg.HistoryRounds)
	if err != nil {
		return CollatorInfo{}, err
	}
	// Get current points
	blocks, err := c.FetchCollatorBlocks(address, c.SnapRound.Number, c.SnapBlock.Hash)
	if err != nil {
		return CollatorInfo{}, err
	}
	// Get amount from pool to avoid bugs in the candidate info data as happened in the past
	pool, err := c.FetchSortedCandidatePool(c.SnapBlock.Hash)
	if err != nil {
		return CollatorInfo{}, err
	}
	counted := TokenAmount{}
	for _, poolEntry := range pool {
		if strings.EqualFold(address, poolEntry.Owner) {
			counted = poolEntry.Amount
			break
		}
	}
	// Get delegations if requested
	cd := make([]DelegatorState, 0)
	if cfg.Revokes {
		var delegations struct {
			Delegations []candidateDelegationUnmarshal
		}
		err = c.GetStorageRaw(
			"ParachainStaking",
			"TopDelegations",
			"Delegations<Balance>",
			&delegations,
			account,
		)
		if err != nil {
			log.Printf("Cannot load delegations %v\n", err)
		}
		// Fetch delegations
		ch := make(chan async.Result[DelegatorState])
		var wg sync.WaitGroup
		for _, delegation := range delegations.Delegations {
			wg.Add(1)
			delegator := delegation.Owner
			go func() {
				defer wg.Done()
				ch <- async.ResultFrom(c.FetchDelegatorState(
					address,
					delegator,
				))
			}()
		}
		// Wait channel
		go func() {
			wg.Wait()
			close(ch)
		}()
		for r := range ch {
			if r.Err != nil {
				log.Printf("Unable to fetch delegator state %v\n", r.Err)
			} else {
				cd = append(cd, r.Value)
			}
		}
		// Sort
		sort.Slice(cd[:], func(i, j int) bool {
			return cd[i].Amount.Balance.Cmp(cd[j].Amount.Balance) == 1
		})
	}
	// Done
	return CollatorInfo{
		Address:     address,
		Selected:    selected,
		Rank:        rank,
		Blocks:      blocks,
		Counted:     counted.AsBalance(&c.TokenInfo),
		MinBond:     candidate.TopAmount.AsBalance(&c.TokenInfo),
		Balance:     info.Balance,
		Display:     info.Identity.Display,
		History:     history,
		Delegations: cd,
	}, nil
}

func (c *Client) FetchCollatorBlocks(address string, round uint32, blockHash types.Hash) (uint32, error) {
	account, _ := types.HexDecodeString(address)
	var roundEncoded = bytes.Buffer{}
	err := scale.NewEncoder(&roundEncoded).Encode(types.NewU32(round))
	if err != nil {
		return 0, err
	}
	// Get points at round
	var points uint32
	err = c.GetStorageRawAt(
		"ParachainStaking",
		"AwardedPts",
		"Points",
		blockHash,
		&points,
		roundEncoded.Bytes(),
		account,
	)
	if err != nil {
		log.Printf("Unable to get points")
		return 0, err
	}
	return points / 20, nil
}

func (c *Client) FetchCollatorHistory(
	address string,
	historyRounds uint32,
) (map[uint32]CollatorHistory, error) {
	account, _ := types.HexDecodeString(address)
	result := make(map[uint32]CollatorHistory, 0)
	for i := c.SnapRound.Number; i >= c.SnapRound.Number-historyRounds; i-- {
		// Cache miss
		blockHash := c.SnapBlock.Hash
		if i < c.SnapRound.Number {
			blockHash, _ = c.GetRoundStartHash(i + 1)
		}
		// Get points at round
		blocks, err := c.FetchCollatorBlocks(address, i, blockHash)
		if err != nil {
			log.Printf("Unable to get points")
			return result, err
		}
		// Now we have the points, lets go back to the end of the round for the rest
		if i < c.SnapRound.Number {
			blockHash, _ = c.GetRoundStartHash(i)
		}
		// Get metadata at round
		var candidate candidateMetadataUnmarshal
		err = c.GetStorageRawAt(
			"ParachainStaking",
			"CandidateInfo",
			"CandidateMetadata<Balance>",
			blockHash,
			&candidate,
			account,
		)
		if err != nil {
			return result, err
		}
		// Get rank at round
		pool, err := c.FetchSortedCandidatePool(blockHash)
		if err != nil {
			return result, err
		}
		rank := getAddressRank(pool, address)
		// Ok
		result[i] = CollatorHistory{
			Blocks:  blocks,
			Rank:    rank,
			Counted: candidate.Counted.AsBalance(&c.TokenInfo),
		}
	}
	return result, nil
}

func (cp *CollatorPool) CollatorInfoByAddress(address string) (CollatorInfo, bool) {
	for _, c := range cp.Collators {
		if strings.EqualFold(address, c.Address) {
			return c, true
		}
	}
	return CollatorInfo{}, false
}

func (cp *CollatorPool) computeRevokes(firstRound uint32, lastRound uint32) {
	tokenInfo := cp.Collators[0].Counted.info
	type sortEntry struct {
		Index   int
		Counted *TokenAmount
		Amount  *TokenAmount
	}
	for round := firstRound; round <= lastRound; round++ {
		roundPool := make([]sortEntry, 0)
		// First we compute new total
		for i, _ := range cp.Collators {
			if cp.Collators[i].Revokes == nil {
				cp.Collators[i].Revokes = make(map[uint32]RevokeRound)
			}
			counted := cp.Collators[i].Counted.Balance.AsBigInt()
			amount := big.NewInt(0)
			if round > firstRound {
				counted = cp.Collators[i].Revokes[round-1].Counted.Balance.AsBigInt()
			}
			for _, delegation := range cp.Collators[i].Delegations {
				if delegation.RevokeRound == round ||
					(round == firstRound && delegation.RevokeRound < round) ||
					(round == lastRound && delegation.RevokeRound > round) {
					counted = counted.Sub(counted, delegation.RevokeAmount.Balance.AsBigInt())
					amount = amount.Add(amount, delegation.RevokeAmount.Balance.AsBigInt())
				}
			}
			// Done
			roundPool = append(
				roundPool,
				sortEntry{Index: i, Counted: &TokenAmount{counted}, Amount: &TokenAmount{amount}},
			)
		}
		// Sort round ranking
		sort.Slice(roundPool[:], func(i, j int) bool {
			return roundPool[i].Counted.Cmp(roundPool[j].Counted) == 1
		})
		// Then we compute new rank
		for rank, r := range roundPool {
			cp.Collators[r.Index].Revokes[round] = RevokeRound{
				Counted: r.Counted.AsBalance(tokenInfo),
				Amount:  r.Amount.AsBalance(tokenInfo),
				Rank:    uint32(rank + 1),
			}
		}
	}
}

func getAddressRank(pool []CandidatePoolEntry, address string) uint32 {
	for i := 0; i < len(pool); i++ {
		if strings.EqualFold(pool[i].Owner, address) {
			return uint32(len(pool) - i)
		}
	}
	return uint32(len(pool))
}
