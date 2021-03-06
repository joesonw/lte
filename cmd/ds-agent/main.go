package main

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/joesonw/lte/cmd/ds-agent/app"
)

func main() {

	var logger *zap.Logger
	rootCmd := &cobra.Command{
		Use:   "distress-ds-agent",
		Short: "distributed stress ds-agent",
	}

	pDebug := rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		var err error
		if *pDebug {
			logger, err = zap.NewDevelopment()
		} else {
			logger, err = zap.NewProduction()
		}
		return err
	}

	rootCmd.AddCommand(app.MakeCmdRun(&logger, pDebug))

	err := rootCmd.Execute()
	if err != nil {
		println(err)
		os.Exit(1)
	}
}
