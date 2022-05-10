package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zooper-corp/mooncli/config"
	"github.com/zooper-corp/mooncli/internal/client"
)

func getClient(cmd *cobra.Command) *client.Client {
	block, _ := cmd.Flags().GetInt64("block")
	round, _ := cmd.Flags().GetUint32("round")
	chain, _ := cmd.Root().Flags().GetString("chain")
	c, err := client.NewClient(config.GetChainConfig(chain, block, round))
	if err != nil {
		panic(err)
	}
	return c
}
