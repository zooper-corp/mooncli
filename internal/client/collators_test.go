package client

import (
	"github.com/zooper-corp/mooncli/config"
	"github.com/zooper-corp/mooncli/internal/display"
	"math/big"
	"strings"
	"testing"
)

func TestClient_FetchSortedCandidatePool(t *testing.T) {
	cfg := config.GetDefaultChainConfig()
	cfg.Snap.TargetBlock = 930124
	c, _ := NewClient(cfg)
	pool, err := c.FetchSortedCandidatePool(c.SnapBlock.Hash)
	if err != nil {
		t.Errorf("error %v\n", err)
	}
	if len(pool) != 77 {
		t.Logf("Client: %v", display.DumpJson(c.SnapBlock))
		t.Logf("Got pool: %v", display.DumpJson(pool))
		t.Errorf("got pool size %v != 76\n", len(pool))
	}
	// Check test collator amount
	e, _ := big.NewInt(0).SetString("1586286471241950000000000", 10)
	for rank, pe := range pool {
		if strings.EqualFold(pe.Owner, config.TestCollatorAddress()) {
			if pe.Amount.Cmp(&TokenAmount{e}) != 0 {
				t.Errorf("invalid amount expected %v, got %v\n", pe.Amount, e)
			}
			if rank != 54 {
				t.Errorf("invalid rank expected %v, got %v\n", 54, rank)
			}
		}
		break
	}
}

func TestClient_GetCandidateBondLessDelay(t *testing.T) {
	cfg := config.GetDefaultChainConfig()
	cfg.Snap.TargetRound = 447
	c, _ := NewClient(cfg)
	delay, err := c.GetCandidateBondLessDelay()
	if delay <= 0 {
		t.Logf("Client: %v", display.DumpJson(c.SnapBlock))
		t.Errorf("got delay %v != 0 err:%v\n", delay, err)
	}
}

func TestClient_FetchCollatorHistory(t *testing.T) {
	cfg := config.GetDefaultChainConfig()
	cfg.Snap.TargetRound = 447
	c, _ := NewClient(cfg)
	history, err := c.FetchCollatorHistory(config.TestCollatorAddress(), 2)
	if err != nil {
		t.Errorf("error %v\n", err)
	}
	if history[446].Blocks == 0 {
		t.Logf("Client: %v", display.DumpJson(c.SnapBlock))
		t.Logf("Go history: %v", display.DumpJson(history))
		t.Errorf("got 0 blocks, wanted > 0 for %v\n", history[1])
	}
}

func TestClient_FetchCollatorInfo(t *testing.T) {
	cfg := config.GetDefaultChainConfig()
	cfg.Snap.TargetBlock = 930124
	c, _ := NewClient(cfg)
	collator, err := c.FetchCollatorInfo(
		config.TestCollatorAddress(),
		true,
		0,
		config.DefaultCollatorsPoolConfig(),
	)
	if err != nil {
		t.Logf("Client: %v", display.DumpJson(c.SnapBlock))
		t.Logf("Go history: %v", display.DumpJson(collator))
		t.Errorf("error %v\n", err)
	}
	e, _ := big.NewInt(0).SetString("1586286471241950000000000", 10)
	if collator.Counted.Balance.Cmp(&TokenAmount{e}) != 0 {
		t.Errorf("invalid counted expected %v, got %v\n", collator.Counted.Balance, e)
	}
	if int64(collator.Counted.Float64()) != 1586286 {
		t.Errorf("invalid counted expected 1586286, got %v\n", int64(collator.Counted.Float64()))
	}
}

func TestClient_FetchRevokes(t *testing.T) {
	cfg := config.GetDefaultChainConfig()
	cfg.Snap.TargetRound = 509
	c, _ := NewClient(cfg)
	poolCfg := config.DefaultCollatorsPoolConfig()
	poolCfg.HistoryRounds = 0
	poolCfg.Revokes = true
	pool, err := c.FetchCollatorPool(poolCfg)
	if err != nil {
		t.Errorf("error %v\n", err)
	}
	collator, ok := pool.CollatorInfoByAddress(config.TestCollatorAddress())
	if !ok {
		t.Errorf("Unable to find collator")
	}
	revoke, ok := collator.Revokes[513]
	if !ok {
		t.Logf("Revokes: %v", display.DumpJson(collator.Revokes))
		t.Errorf("Expecting revokes at round 500")
	}
	if revoke.Amount.Float64() != 2500.0 {
		t.Logf("Revoke: %v", display.DumpJson(revoke))
		t.Errorf("Expecting 700 got: %v", collator.Revokes[500].Amount.Balance)
	}
}
