package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zooper-corp/mooncli/config"
	"github.com/zooper-corp/mooncli/internal/display"
	"github.com/zooper-corp/mooncli/internal/server"
	"log"
	"time"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve collator and delegator data JSON via API server",
	Run: func(cmd *cobra.Command, args []string) {
		chain, _ := cmd.Root().Flags().GetString("chain")
		interval, _ := cmd.Flags().GetUint32("interval")
		listen, _ := cmd.Flags().GetString("listen")
		httpConfig := config.HttpConfig{
			Addr:           listen,
			UpdateInterval: time.Duration(interval) * time.Second,
			ChainConfig:    config.GetChainConfig(chain, 0, 0),
		}
		log.Printf("Starting API server %v", display.DumpJson(httpConfig))
		server.ServeChainData(httpConfig)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	httpConfig := config.GetDefaultHttpConfig()
	serveCmd.PersistentFlags().Uint32(
		"interval",
		uint32(httpConfig.UpdateInterval.Seconds()),
		"Max update interval",
	)
	serveCmd.PersistentFlags().String(
		"listen",
		httpConfig.Addr,
		"Listen address",
	)
}
