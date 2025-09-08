package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nanthakumaran-s/golancer/internal/config"
	"github.com/nanthakumaran-s/golancer/internal/runway"
	"github.com/nanthakumaran-s/golancer/internal/tower"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var addr string
var configFile string

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&addr, "port", "p", "8080", "Port for Golancer")
	startCmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Path to configuration file")

	_ = viper.BindPFlag("port", startCmd.Flags().Lookup("port"))
	_ = viper.BindPFlag("config", startCmd.Flags().Lookup("config"))
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the golancer server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		port := viper.GetInt("port")

		imm := &config.ServerDefaults{
			Port: port,
		}

		dp := runway.NewDataPlane(imm)
		cp := tower.NewControlPlane(dp)
		cp.Start(ctx)

		select {
		case sig := <-sigCh:
			fmt.Println("Received signal:", sig)
			cancel()
			cp.Stop()
		}

		return nil
	},
}
