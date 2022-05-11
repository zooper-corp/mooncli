package client

import (
	"encoding/json"
	"fmt"
	"github.com/OrlovEvgeny/go-mcache"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/client"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	scalecodec "github.com/itering/scale.go"
	"github.com/itering/scale.go/source"
	types2 "github.com/itering/scale.go/types"
	"github.com/itering/scale.go/utiles"
	"github.com/zooper-corp/mooncli/config"
	"log"
	"time"
)

type Client struct {
	api         *gsrpc.SubstrateAPI
	cache       *mcache.CacheDriver
	metadata    *types.Metadata
	decoder     scalecodec.MetadataDecoder
	RpcUrl      string      `json:"endpoint"`
	Chain       string      `json:"chain"`
	SpecVersion int         `json:"spec"`
	SnapBlock   SnapBlock   `json:"block"`
	SnapRound   SnapRound   `json:"round"`
	SnapStaking SnapStaking `json:"candidate_pool"`
	TokenInfo   TokenInfo   `json:"token"`
}

func NewClient(config config.ChainConfig) (*Client, error) {
	return NewClientWithExternalCache(config, mcache.New())
}

func NewClientWithExternalCache(cfg config.ChainConfig, cache *mcache.CacheDriver) (*Client, error) {
	c := new(Client)
	c.cache = cache
	c.RpcUrl = cfg.RpcUrl()
	// Create client first
	api, err := gsrpc.NewSubstrateAPI(c.RpcUrl)
	if err != nil {
		return c, err
	}
	c.api = api
	// Get metadata
	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return c, err
	}
	c.metadata = meta
	// Load generic decoder first
	err = c.registerDecoder(cfg)
	if err != nil {
		return c, err
	}
	// Get snap
	snap, err := fetchSnapBlock(c, cfg.Snap.TargetBlock, cfg.Snap.TargetRound)
	if err != nil {
		return c, err
	}
	c.SnapBlock = snap.Block
	c.SnapRound = snap.Round
	c.SnapStaking = snap.Staking
	// Get chain info
	chain, err := api.RPC.System.Chain()
	if err != nil {
		return c, err
	}
	c.Chain = string(chain)
	// Get version at snap point
	version, _ := api.RPC.State.GetRuntimeVersion(c.SnapBlock.Hash)
	if err != nil {
		return c, err
	}
	c.SpecVersion = int(version.SpecVersion)
	// Reload decoder
	cfg.NetworkSpecsVersion = uint32(version.SpecVersion)
	err = c.registerDecoder(cfg)
	if err != nil {
		return c, err
	}
	// Get token info
	tokenInfo, err := fetchTokenInfo(c)
	if err != nil {
		return c, err
	}
	c.TokenInfo = tokenInfo
	// Done
	log.Printf(
		"Connected to %v v%v at block %v hash %v\n",
		c.Chain,
		c.SpecVersion,
		c.SnapBlock.Number,
		c.SnapBlock.Hash.Hex(),
	)
	return c, nil
}

func (c *Client) registerDecoder(cfg config.ChainConfig) error {
	var hexMetadata string
	err := client.CallWithBlockHash(c.api.Client, &hexMetadata, "state_getMetadata", nil)
	if err != nil {
		return err
	}
	metaDecoder := scalecodec.MetadataDecoder{}
	metaDecoder.Init(utiles.HexToBytes(hexMetadata))
	_ = metaDecoder.Process()
	customType, err := cfg.ReadSpecs()
	if err != nil {
		return err
	}
	types2.RegCustomTypes(source.LoadTypeRegistry(customType))
	c.decoder = metaDecoder
	return nil
}

func (c *Client) GetBlockNumber(hash types.Hash) (uint64, error) {
	headerLatest, err := c.api.RPC.Chain.GetHeader(hash)
	if err != nil {
		return 0, err
	}
	return uint64(headerLatest.Number), nil
}

func (c *Client) GetRoundStartHash(round uint32) (types.Hash, error) {
	roundDelta := c.SnapRound.Number - round
	blockDelta := roundDelta * c.SnapRound.Length
	startBlock := c.SnapRound.Start - blockDelta
	return c.api.RPC.Chain.GetBlockHash(uint64(startBlock))
}

func (c *Client) GetRoundEndHash(round uint32) (types.Hash, error) {
	roundDelta := c.SnapRound.Number - round
	blockDelta := roundDelta * c.SnapRound.Length
	startBlock := c.SnapRound.Start - blockDelta
	return c.api.RPC.Chain.GetBlockHash(uint64(startBlock + c.SnapRound.Length - 2))
}

func (c *Client) GetStorage(
	pallet string,
	method string,
	target interface{},
	args ...[]byte,
) (ok bool, err error) {
	return c.GetStorageAt(pallet, method, target, c.SnapBlock.Hash, args...)
}

func (c *Client) GetStorageAt(
	pallet string,
	method string,
	target interface{},
	blockHash types.Hash,
	args ...[]byte,
) (ok bool, err error) {
	key, err := types.CreateStorageKey(c.metadata, pallet, method, args...)
	if err != nil {
		return false, err
	}
	return c.api.RPC.State.GetStorage(key, target, blockHash)
}

// GetStorageRaw will fetch storage at client snap block with minimum cache TTL (since we have no block reference)
func (c *Client) GetStorageRaw(
	pallet string,
	method string,
	typeString string,
	targetValue any,
	args ...[]byte,
) error {
	return c.GetStorageRawWithTtl(pallet, method, typeString, config.MinCacheTTL(), targetValue, args...)
}

// GetStorageRawWithTtl will fetch storage at client snap block caching with given TTL and ignoring block
func (c *Client) GetStorageRawWithTtl(
	pallet string,
	method string,
	typeString string,
	cacheTtl time.Duration,
	targetValue any,
	args ...[]byte,
) error {
	cacheKey := fmt.Sprintf("%v.%v(%v)", pallet, method, args)
	return c.getStorage(pallet, method, typeString, c.SnapBlock.Hash, cacheTtl, cacheKey, targetValue, args...)
}

// GetStorageRawAt will fetch storage at given block with default cache TTL
func (c *Client) GetStorageRawAt(
	pallet string,
	method string,
	typeString string,
	blockHash types.Hash,
	targetValue any,
	args ...[]byte,
) error {
	cacheTtl := config.DefaultCacheTTL()
	cacheKey := fmt.Sprintf("%v.%v(%v)@%v", pallet, method, args, blockHash)
	return c.getStorage(pallet, method, typeString, blockHash, cacheTtl, cacheKey, targetValue, args...)
}

// GetConstantValue will fetch a constant value and Marshal it as JSON
func (c *Client) GetConstantValue(
	pallet string,
	method string,
	typeString string,
	targetValue any,
) error {
	raw, err := c.metadata.FindConstantValue(pallet, method)
	if err != nil {
		return err
	}
	j, err := c.decodeRawData(raw, typeString)
	if err != nil {
		return err
	}
	err = json.Unmarshal(j, targetValue)
	if err != nil {
		return err
	}
	return nil
}

// getStorage will fetch storage at given block with given cache ttl and key
func (c *Client) getStorage(
	pallet string,
	method string,
	typeString string,
	blockHash types.Hash,
	cacheTtl time.Duration,
	cacheKey string,
	targetValue any,
	args ...[]byte,
) error {
	cache, ok := c.getCache(cacheKey)
	if ok {
		err := json.Unmarshal(cache.([]byte), targetValue)
		if err == nil {
			return nil
		}
	}
	r, err := c.getStorageData(pallet, method, blockHash, args...)
	if err != nil {
		return err
	}
	// Decode
	j, err := c.decodeRawData(r, typeString)
	if err != nil {
		return err
	}
	// Cache
	c.setCache(cacheKey, j, cacheTtl)
	err = json.Unmarshal(j, targetValue)
	if err != nil {
		return err
	}
	return nil
}

// getStorageData will fetch storage raw data at given block
func (c *Client) getStorageData(
	pallet string,
	method string,
	blockHash types.Hash,
	args ...[]byte,
) (data []byte, err error) {
	key, err := types.CreateStorageKey(c.metadata, pallet, method, args...)
	if err != nil {
		return nil, err
	}
	raw, err := c.api.RPC.State.GetStorageRaw(key, blockHash)
	if err != nil {
		return nil, err
	}
	// Unmarshal
	if err != nil {
		return nil, err
	}
	return *raw, err
}

// decodeRawData will decode given raw data as typeString
func (c *Client) decodeRawData(raw []byte, typeString string) ([]byte, error) {
	decoder := types2.ScaleDecoder{}
	option := types2.ScaleDecoderOption{Metadata: &c.decoder.Metadata}
	decoder.Init(types2.ScaleBytes{Data: raw}, &option)
	r := decoder.ProcessAndUpdateData(typeString)
	// Marshal in JSON
	j, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// getCache - returns serialize data
func (c *Client) getCache(key string) (interface{}, bool) {
	r, ok := c.cache.Get(key)
	if ok && r != nil {
		return r, ok
	} else {
		return r, false
	}
}

// setCache - add cache data value
func (c *Client) setCache(key string, value interface{}, ttl time.Duration) {
	err := c.cache.Set(key, value, ttl)
	if err != nil {
		log.Printf("Unable to write cache %v", err)
	}
}
