package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zooper-corp/mooncli/config"
	"github.com/zooper-corp/mooncli/internal/async"
	"github.com/zooper-corp/mooncli/internal/client"
	"github.com/zooper-corp/mooncli/internal/tools"
	"log"
	"sync"
)

type infoResult struct {
	Metadata *client.Client       `json:"info"`
	Accounts []client.AccountInfo `json:"accounts,omitempty"`
}

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show chain info at head or a specific block / round",
	Run: func(cmd *cobra.Command, args []string) {
		block, _ := cmd.Flags().GetInt64("block")
		round, _ := cmd.Flags().GetUint32("round")
		chain, _ := cmd.Root().Flags().GetString("chain")
		c, err := client.NewClient(config.GetChainConfig(chain, block, round))
		if err != nil {
			panic(err)
		}
		// Fetch address info if there
		addresses, _ := cmd.Flags().GetStringSlice("address")
		log.Printf("Fetching account info for %v\n", addresses)
		ch := make(chan async.Result[client.AccountInfo])
		var wg sync.WaitGroup
		go func() {
			for _, address := range addresses {
				wg.Add(1)
				go fetchAccountInfo(ch, &wg, c, address)
			}
			wg.Wait()
			close(ch)
		}()
		var accounts []client.AccountInfo
		for r := range ch {
			if r.Err != nil {
				log.Printf("Unable to fetch address info %v\n", r.Err)
			} else {
				accounts = append(accounts, r.Value)
			}
		}
		// Build result
		result := infoResult{
			Metadata: c,
			Accounts: accounts,
		}
		fmt.Println(tools.DumpJson(result))
	},
}

func fetchAccountInfo(
	ch chan async.Result[client.AccountInfo],
	wg *sync.WaitGroup,
	c *client.Client,
	address string,
) {
	defer wg.Done()
	account, err := c.FetchAccountInfo(address)
	if err != nil {
		ch <- async.ErrorResult[client.AccountInfo](err)
		return
	}
	ch <- async.SuccessResult[client.AccountInfo](account)
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.PersistentFlags().Int64(
		"block",
		0,
		"Absolute block or position relative to the round",
	)
	infoCmd.PersistentFlags().Uint32(
		"round",
		0,
		"Round number, when used block will be relative",
	)
	infoCmd.PersistentFlags().StringSlice(
		"address",
		[]string{},
		"Also retrieves info on given addresses",
	)
}
