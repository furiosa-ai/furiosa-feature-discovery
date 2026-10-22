package cmd

import (
	"github.com/furiosa-ai/furiosa-feature-discovery/internal/config"
	"github.com/furiosa-ai/furiosa-feature-discovery/internal/consts"
	"github.com/furiosa-ai/furiosa-feature-discovery/internal/discovery"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var updateInterval int
	var outputPath string

	cmd := &cobra.Command{
		Use:          "furiosa-feature-discovery",
		Short:        "Furiosa Feature Discovery",
		Args:         cobra.MaximumNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.NewConfig(updateInterval, outputPath)
			runner, err := discovery.NewFeatureDiscoveryRunner(cfg)
			if err != nil {
				return err
			}

			return runner.Run()
		},
	}

	cmd.Flags().IntVar(&updateInterval, "interval", consts.DefaultLabelUpdateIntervalSeconds, "Update interval in seconds")
	cmd.Flags().StringVar(&outputPath, "output", consts.DefaultOutputPath, "Output file path")

	return cmd
}
