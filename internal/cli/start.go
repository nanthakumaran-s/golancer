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

var (
	addr       int
	configFile string
	useTLS     bool
	local      bool
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVarP(&addr, "port", "p", 8080, "Port for Golancer")
	startCmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Path to configuration file")
	startCmd.Flags().BoolVarP(&useTLS, "useTLS", "", false, "Set Golancer to use TLS")
	startCmd.Flags().BoolVarP(&local, "local", "", false, "Set Golancer in local development mode")

	_ = viper.BindPFlag("port", startCmd.Flags().Lookup("port"))
	_ = viper.BindPFlag("config", startCmd.Flags().Lookup("config"))
	_ = viper.BindPFlag("useTLS", startCmd.Flags().Lookup("useTLS"))
	_ = viper.BindPFlag("local", startCmd.Flags().Lookup("local"))
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the golancer server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		imm := &config.ServerDefaults{
			Port:   addr,
			UseTLS: useTLS,
			Local:  local,
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
