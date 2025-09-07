package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "golancer",
	Short: "GoLancer is a modern reverse proxy & load balancer",
	Long:  `GoLancer is a lightweight reverse proxy and load balancer with YAML config, observability, and security features.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
