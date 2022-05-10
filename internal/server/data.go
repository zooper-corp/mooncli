package server

import (
	"fmt"
	"github.com/OrlovEvgeny/go-mcache"
	"github.com/zooper-corp/mooncli/config"
	"github.com/zooper-corp/mooncli/internal/client"
	"log"
	"strings"
	"sync"
	"time"
)

type ChainData struct {
	cache          *mcache.CacheDriver
	dataLock       sync.RWMutex
	updateLock     sync.Mutex
	chainConfig    config.ChainConfig
	maxUpdateDelta time.Duration
	Info           ChainInfo             `json:"info"`
	Collators      []client.CollatorInfo `json:"collators"`
}

type CollatorData struct {
	Info      ChainInfo             `json:"info"`
	Collators []client.CollatorInfo `json:"collators"`
}

type DelegationInfo struct {
	Collator     string              `json:"collator"`
	Address      string              `json:"address"`
	Amount       client.TokenBalance `json:"amount"`
	RevokeAmount client.TokenBalance `json:"revoke_amount,omitempty"`
	RevokeReason string              `json:"revoke_reason,omitempty"`
	RevokeRound  uint32              `json:"revoke_round,omitempty"`
}

type DelegationData struct {
	Info        ChainInfo        `json:"info"`
	Delegations []DelegationInfo `json:"delegations"`
}

type ChainInfo struct {
	Server      string             `json:"server"`
	Update      ChainUpdate        `json:"update"`
	Chain       string             `json:"chain"`
	SpecVersion int                `json:"spec"`
	SnapBlock   client.SnapBlock   `json:"block"`
	SnapRound   client.SnapRound   `json:"round"`
	SnapStaking client.SnapStaking `json:"candidate_pool"`
	TokenInfo   client.TokenInfo   `json:"token"`
}

type ChainUpdate struct {
	TsSecs  float64 `json:"ts"`
	LenSecs float32 `json:"len"`
}

func NewChainData(chainConfig config.ChainConfig, maxUpdateDelta time.Duration) (ChainData, error) {
	return ChainData{
		cache:          mcache.New(),
		dataLock:       sync.RWMutex{},
		chainConfig:    chainConfig,
		maxUpdateDelta: maxUpdateDelta,
	}, nil
}

func (c *ChainData) Update() error {
	if c.updateLock.TryLock() {
		start := time.Now().UnixMilli()
		// Create basic client
		log.Printf("Starting update")
		chainClient, err := client.NewClientWithExternalCache(c.chainConfig, c.cache)
		if err != nil {
			log.Printf("Unable to create client %v", err)
			return err
		}
		// Fetch collator pool
		log.Printf("Fetching collator pool")
		collatorPool, err := chainClient.FetchCollatorPool(config.CollatorsPoolConfig{
			HistoryRounds: 28,
			Revokes:       true,
		})
		if err != nil {
			log.Printf("Unable to fetch collator pool %v", err)
			return err
		}
		// Check pool size
		if len(collatorPool.Collators) != int(chainClient.SnapStaking.Total) {
			log.Printf(
				"Fetched pool size is %v expecting %v, not updating",
				len(collatorPool.Collators),
				int(chainClient.SnapStaking.Total),
			)
			return fmt.Errorf("pool size does not match")
		}
		// Done update backend
		c.dataLock.Lock()
		defer c.dataLock.Unlock()
		defer c.updateLock.Unlock()
		updateTime := uint32(time.Now().UnixMilli() - start)
		c.Info = ChainInfo{
			Server: "MoonCli by 🛸 Zooper Corp 🛸",
			Update: ChainUpdate{
				TsSecs:  float64(start / 1000),
				LenSecs: float32(updateTime / 1000),
			},
			Chain:       chainClient.Chain,
			SpecVersion: 0,
			SnapBlock:   chainClient.SnapBlock,
			SnapRound:   chainClient.SnapRound,
			SnapStaking: chainClient.SnapStaking,
			TokenInfo:   chainClient.TokenInfo,
		}
		c.Collators = collatorPool.Collators
		// Finished
		log.Printf("Chain update done in %vs", float64(updateTime*100)/100000.0)
		return nil
	} else {
		log.Printf("Chain update already in progress")
		return nil
	}
}

func (c *ChainData) GetInfo() *ChainInfo {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	chainInfo := c.Info
	return &chainInfo
}

func (c *ChainData) GetCollators() CollatorData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	chainData := c
	return CollatorData{
		Info:      chainData.Info,
		Collators: chainData.Collators,
	}
}

func (c *ChainData) GetDelegations(address string) DelegationData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	chainData := c
	result := make([]DelegationInfo, 0)
	for _, collator := range chainData.Collators {
		for _, delegation := range collator.Delegations {
			if strings.EqualFold(collator.Address, address) || strings.EqualFold(delegation.Address, address) {
				result = append(result, DelegationInfo{
					Collator:     collator.Address,
					Address:      delegation.Address,
					Amount:       delegation.Amount,
					RevokeRound:  delegation.RevokeRound,
					RevokeAmount: delegation.RevokeAmount,
					RevokeReason: delegation.RevokeReason,
				})
			}
		}
	}
	return DelegationData{
		Info:        chainData.Info,
		Delegations: result,
	}
}

func (c *ChainData) GetCollator(address string) CollatorData {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	chainData := c
	filtered := make([]client.CollatorInfo, 0)
	for _, collator := range chainData.Collators {
		if strings.EqualFold(collator.Address, address) {
			filtered = append(filtered, collator)
		}
	}
	return CollatorData{
		Info:      chainData.Info,
		Collators: filtered,
	}
}
