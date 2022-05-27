package config

import (
	"embed"
	"fmt"
	"math/rand"
	"time"
)

// content holds our static web server content.
//go:embed specs/*
var networkSpecs embed.FS

type ChainConfig struct {
	Endpoints []string
	// Snap point
	Snap SnapConfig
	// Timeouts
	DialTimeout      time.Duration
	SubscribeTimeout time.Duration
	// Json folders
	NetworkSpecs        string
	NetworkSpecsVersion uint32
}

type SnapConfig struct {
	TargetBlock int64
	TargetRound uint32
}

// MinCacheTTL Duration the minimum time an object is valid, this is usually 2 * block length
func MinCacheTTL() time.Duration {
	return 24 * time.Second
}

// DefaultCacheTTL Duration the default time an object is valid (this expects object to be requested with hash)
func DefaultCacheTTL() time.Duration {
	return 24 * time.Hour
}

// GetDefaultChainConfig ChainConfig default values, main network
func GetDefaultChainConfig() ChainConfig {
	return GetChainConfig("moonbeam", 0, 0)
}

// GetChainConfig ChainConfig returns the default config for a given network or the default network
func GetChainConfig(
	endpoint string,
	block int64,
	round uint32,
) ChainConfig {
	endpoints := extractDefaultRpcUrl(endpoint)
	return ChainConfig{
		Endpoints: endpoints,
		Snap: SnapConfig{
			TargetBlock: block,
			TargetRound: round,
		},
		DialTimeout:         10 * time.Second,
		SubscribeTimeout:    5 * time.Second,
		NetworkSpecs:        "moonbeam.1502",
		NetworkSpecsVersion: 1502,
	}
}

// ReadSpecs will read network specification from the embedded file
func (cg *ChainConfig) ReadSpecs() ([]byte, error) {
	return networkSpecs.ReadFile(fmt.Sprintf("specs/%v.json", cg.NetworkSpecs))
}

// RpcUrl will return a random RPC address to connect to
func (cg *ChainConfig) RpcUrl() string {
	rand.Seed(time.Now().Unix())
	return cg.Endpoints[rand.Intn(len(cg.Endpoints))]
}

// TestCollatorAddress returns string address of collator used for testing (Foundation 04)
func TestCollatorAddress() string {
	return "0xf02ddb48eda520c915c0dabadc70ba12d1b49ad2"
}

// ExtractDefaultRPCURL reads the env variable RPC_URL and returns it. If that variable is unset or empty,
// it will fallback to "http://127.0.0.1:9933"
func extractDefaultRpcUrl(endpoint string) []string {
	switch endpoint {
	case "moonbeam":
		return []string{
			"wss://moonbeam.api.onfinality.io/public-ws",
			"wss://wss.api.moonbeam.network",
		}
	case "moonriver":
		return []string{
			"wss://moonriver.api.onfinality.io/public-ws",
			"wss://wss.api.moonriver.moonbeam.network",
		}
	case "moonbase":
		return []string{
			"wss://moonbeam-alpha.api.onfinality.io/public-ws",
			"wss://wss.api.moonbase.moonbeam.network",
		}
	default:
		return []string{endpoint}
	}
}
