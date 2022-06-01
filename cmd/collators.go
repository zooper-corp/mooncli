package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zooper-corp/mooncli/config"
	"github.com/zooper-corp/mooncli/internal/client"
	"github.com/zooper-corp/mooncli/internal/display"
	"github.com/zooper-corp/mooncli/internal/tools"
	"log"
)

// collatorsCmd represents the collators command
var collatorsCmd = &cobra.Command{
	Use:   "collators",
	Short: "Shows collator pool statistics",
}

// collatorsTableCmd represents the collators command
var collatorsTableCmd = &cobra.Command{
	Use:   "table",
	Short: "Shows collator pool statistics as table",
	Run: func(cmd *cobra.Command, args []string) {
		compact, _ := cmd.Flags().GetBool("compact")
		sortKey, _ := cmd.Flags().GetString("sort-key")
		sortDesc, _ := cmd.Flags().GetBool("sort-desc")
		revokeRounds, _ := cmd.Flags().GetUint32("revoke-rounds")
		options := config.TableOptions{
			Compact:      compact,
			SortKey:      sortKey,
			SortDesc:     sortDesc,
			RevokeRounds: revokeRounds,
		}
		data, client := fetchPool(cmd)
		display.DumpTable(data, client, options)
	},
}

// collatorsJsonCmd represents the collators command
var collatorsJsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Dumps collator pool statistics as json",
	Run: func(cmd *cobra.Command, args []string) {
		type JsonData struct {
			Client *client.Client      `json:"info"`
			Pool   client.CollatorPool `json:"collator_pool"`
		}
		data, client := fetchPool(cmd)
		fmt.Println(tools.DumpJson(JsonData{Client: client, Pool: data}))
	},
}

func fetchPool(cmd *cobra.Command) (client.CollatorPool, *client.Client) {
	c := getClient(cmd)
	// Get the pool
	historyRounds, _ := cmd.Flags().GetUint32("history")
	revokes, _ := cmd.Flags().GetBool("revokes")
	address, _ := cmd.Flags().GetString("address")
	log.Printf("Fetching collator pool history:%v revokes:%v\n", historyRounds, revokes)
	data, err := c.FetchCollatorPool(config.CollatorsPoolConfig{
		Address:       address,
		HistoryRounds: historyRounds,
		Revokes:       revokes,
	})
	if err != nil {
		panic(err)
	}
	return data, c
}

func init() {
	rootCmd.AddCommand(collatorsCmd)
	collatorsPoolConfig := config.DefaultCollatorsPoolConfig()
	collatorsCmd.PersistentFlags().Int64(
		"block",
		0,
		"Absolute block or position relative to the round",
	)
	collatorsCmd.PersistentFlags().Uint32(
		"round",
		0,
		"Round number, when used block will be relative",
	)
	collatorsCmd.PersistentFlags().Uint32(
		"history",
		collatorsPoolConfig.HistoryRounds,
		"Number of rounds to fetch points history",
	)
	collatorsCmd.PersistentFlags().Bool(
		"revokes",
		collatorsPoolConfig.Revokes,
		"Fetch delegations and revokes for collator",
	)
	collatorsCmd.PersistentFlags().String(
		"address",
		collatorsPoolConfig.Address,
		"Retrieve info only for a given address",
	)
	collatorsCmd.AddCommand(collatorsTableCmd)
	collatorsTableCmd.PersistentFlags().Bool(
		"compact",
		config.GetDefaultTableOptions().Compact,
		"Shows compact table",
	)
	collatorsTableCmd.PersistentFlags().String(
		"sort-key",
		config.GetDefaultTableOptions().SortKey,
		"Sort table by key",
	)
	collatorsTableCmd.PersistentFlags().Bool(
		"sort-desc",
		config.GetDefaultTableOptions().SortDesc,
		"Sort table in descending order",
	)
	collatorsTableCmd.PersistentFlags().Uint32(
		"revoke-rounds",
		config.GetDefaultTableOptions().RevokeRounds,
		"Number of rounds to show in revoke stats",
	)
	collatorsCmd.AddCommand(collatorsJsonCmd)
}
