package client

import (
	"github.com/zooper-corp/mooncli/config"
	"github.com/zooper-corp/mooncli/internal/tools"
	"testing"
)

func TestFetchDelegatorInfo(t *testing.T) {
	cfg := config.GetDefaultChainConfig()
	cfg.Snap.TargetRound = 447
	c, _ := NewClient(cfg)
	delegator, err := c.FetchDelegatorState(
		"0xca98d4378393040408100f490bf98b03f5e7deb7",
		"0x4b5788f50e44e593c7bd92eb66fa59600baa9432",
	)
	if err != nil {
		t.Logf("Client: %v", tools.DumpJson(c.SnapBlock))
		t.Logf("Go history: %v", tools.DumpJson(delegator))
		t.Errorf("error %v\n", err)
	}
}

func TestClient_FetchDelegatorState(t *testing.T) {
	cfg := config.GetDefaultChainConfig()
	cfg.Snap.TargetBlock = 934930
	c, _ := NewClient(cfg)
	delegator, err := c.FetchDelegatorState(
		"0x3f0937BdEF510fd1D39F76CF41a7A4CFbf8ab876",
		"0x728507eC8f967BCB5fAFF3D238059cE1eb99b828",
	)
	if err != nil {
		t.Logf("Client: %v", tools.DumpJson(c.SnapBlock))
		t.Logf("Go delegator: %v", tools.DumpJson(delegator))
		t.Errorf("error %v\n", err)
	}
}

func TestClient_FetchDelegatorStateV1500(t *testing.T) {
	cfg := config.GetChainConfig("moonbase", 2132368, 0)
	cfg.Snap.TargetBlock = 2132368
	c, _ := NewClient(cfg)
	delegator, err := c.FetchDelegatorState(
		"0x7aF6c67EE0F1eC83C3d05e62fB0200B3841c7F36",
		"0xB1e6c73EA591C3d1cE112f428B33850E7158fe22",
	)
	if err != nil {
		t.Logf("Client: %v", tools.DumpJson(c.SnapBlock))
		t.Logf("Go delegator: %v", tools.DumpJson(delegator))
		t.Errorf("error %v\n", err)
	}
	if delegator.RevokeRound != 2885 {
		t.Logf("Client: %v", tools.DumpJson(c.SnapBlock))
		t.Logf("Go history: %v", tools.DumpJson(delegator))
		t.Errorf("expected revoke round 2885 got %v\n", delegator.RevokeRound)
	}
}

/*
{
{
  id: 0x4b5788F50E44e593C7Bd92eB66fa59600bAA9432
  delegations: [
    {
      owner: 0x0a0952E7d58817C40473D57a7E37f188DdB81ff9
      amount: 4,590,000,000,000,000,000,000
    }
    {
      owner: 0x564E8464A616baE3c366467eD572C3d2Ae8b9E63
      amount: 4,500,000,000,000,000,000,000
    }
    {
      owner: 0xCA98D4378393040408100f490bF98b03F5E7DeB7
      amount: 10,448,000,000,000,000,000,000
    }
    {
      owner: 0xeCca07badBd38937122B82ec8AfCf86b1E2b7939
      amount: 11,576,000,000,000,000,000,000
    }
  ]
  total: 31,114,000,000,000,000,000,000
  requests: {
    revocationsCount: 3
    requests: {
      0x0a0952E7d58817C40473D57a7E37f188DdB81ff9: {
        collator: 0x0a0952E7d58817C40473D57a7E37f188DdB81ff9
        amount: 4,590,000,000,000,000,000,000
        whenExecutable: 454
        action: Revoke
      }
      0x564E8464A616baE3c366467eD572C3d2Ae8b9E63: {
        collator: 0x564E8464A616baE3c366467eD572C3d2Ae8b9E63
        amount: 4,500,000,000,000,000,000,000
        whenExecutable: 454
        action: Revoke
      }
      0xCA98D4378393040408100f490bF98b03F5E7DeB7: {
        collator: 0xCA98D4378393040408100f490bF98b03F5E7DeB7
        amount: 10,448,000,000,000,000,000,000
        whenExecutable: 462
        action: Revoke
      }
    }
    lessTotal: 19,538,000,000,000,000,000,000
  }
  status: Active
}
*/
