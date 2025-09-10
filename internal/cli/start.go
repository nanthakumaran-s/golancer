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
	"github.com/nanthakumaran-s/golancer/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	addr       int
	configFile string
	useTLS     bool
	local      bool
	logFile    string
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVarP(&addr, utils.PORT, "p", 8080, "Port for Golancer")
	startCmd.Flags().StringVarP(&configFile, utils.CONFIG, "c", "config.yaml", "Path to configuration file")
	startCmd.Flags().BoolVarP(&useTLS, utils.UseTLS, "", false, "Set Golancer to use TLS")
	startCmd.Flags().BoolVarP(&local, utils.Local, "", false, "Set Golancer in local development mode")
	startCmd.Flags().StringVarP(&logFile, utils.LogFile, "l", "golancer.log", "Set Golancer logging file")

	_ = viper.BindPFlag(utils.PORT, startCmd.Flags().Lookup(utils.PORT))
	_ = viper.BindPFlag(utils.CONFIG, startCmd.Flags().Lookup(utils.CONFIG))
	_ = viper.BindPFlag(utils.UseTLS, startCmd.Flags().Lookup(utils.UseTLS))
	_ = viper.BindPFlag(utils.Local, startCmd.Flags().Lookup(utils.Local))
	_ = viper.BindPFlag(utils.LogFile, startCmd.Flags().Lookup(utils.LogFile))
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
