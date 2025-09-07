package cli

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/nanthakumaran-s/golancer/internal/config"
	"github.com/nanthakumaran-s/golancer/internal/server"
	"github.com/spf13/cobra"
)

var configFile string

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Path to configuration file")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the golancer server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgMgr, err := config.NewManager(configFile)
		if err != nil {
			return err
		}

		server := server.NewServer(cfgMgr.Get())

		updates := cfgMgr.Subscribe()
		go func() {
			for newCfg := range updates {
				fmt.Println("[server] applying new config...")
				server.ApplyConfig(newCfg)
			}
		}()

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		go func() {
			if err := server.Start(); err != nil {
				fmt.Println("server stopped:", err)
			}
		}()

		<-ctx.Done()
		fmt.Println("shutdown signal received")
		return nil
	},
}
