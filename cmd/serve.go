package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zooper-corp/mooncli/config"
	"github.com/zooper-corp/mooncli/internal/server"
	"github.com/zooper-corp/mooncli/internal/tools"
	"log"
	"runtime"
	"time"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve collator and delegator data JSON via API server",
	Run: func(cmd *cobra.Command, args []string) {
		runtime.GOMAXPROCS(1)
		chain, _ := cmd.Root().Flags().GetString("chain")
		interval, _ := cmd.Flags().GetUint32("interval")
		listen, _ := cmd.Flags().GetString("listen")
		dataPath, _ := cmd.Flags().GetString("data-path")
		httpConfig := config.HttpConfig{
			Addr:           listen,
			UpdateInterval: time.Duration(interval) * time.Second,
			ChainConfig:    config.GetChainConfig(chain, 0, 0),
			DataPath:       dataPath,
		}
		log.Printf("Starting API server %v", tools.DumpJson(httpConfig))
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
	serveCmd.PersistentFlags().String(
		"data-path",
		"",
		"An optional data path, if provided update data will be cached there",
	)
}
