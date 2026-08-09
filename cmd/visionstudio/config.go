package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	config "github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage visionstudio configuration",
		Long: `View and modify the visionstudio configuration file.

Configuration is stored at ~/.productbuildershq/visionstudio/config.json.
Values are resolved in order: CLI flag > environment variable > config file > default.`,
	}
	cmd.AddCommand(
		configShowCmd(),
		configSetCmd(),
		configUnsetCmd(),
		configPathCmd(),
	)
	return cmd
}

func configShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			cmd.Printf("Config file: %s\n\n", path)

			hasValues := false
			if cfg.DSN != "" {
				cmd.Printf("dsn = %s\n", cfg.DSN)
				hasValues = true
			}
			if cfg.Defaults.Workflow != "" {
				cmd.Printf("defaults.workflow = %s\n", cfg.Defaults.Workflow)
				hasValues = true
			}
			if !hasValues {
				cmd.Println("(no values set)")
			}

			cmd.Println()
			cmd.Println("Resolved DSN:")
			if flag, _ := cmd.Root().Flags().GetString("dsn"); flag != "" {
				cmd.Printf("  %s  (from --dsn flag)\n", flag)
			} else if env := os.Getenv("VISIONSTUDIO_DSN"); env != "" {
				cmd.Printf("  %s  (from $VISIONSTUDIO_DSN)\n", env)
			} else if cfg.DSN != "" {
				cmd.Printf("  %s  (from config file)\n", cfg.DSN)
			} else {
				cmd.Printf("  (embedded mode — no DSN configured)\n")
			}
			return nil
		},
	}
}

func configSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a configuration value. Supported keys:
  dsn                MySQL-compatible DSN for Dolt server mode
  defaults.workflow  Default spec workflow ID (e.g. pbhq-lite, aws-product)`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			switch key {
			case "dsn":
				cfg.DSN = value
			case "defaults.workflow":
				cfg.Defaults.Workflow = value
			default:
				return fmt.Errorf("unknown config key: %s (supported: dsn, defaults.workflow)", key)
			}

			if err := cfg.Save(); err != nil {
				return err
			}
			cmd.Printf("Set %s = %s\n", key, value)
			return nil
		},
	}
}

func configUnsetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unset <key>",
		Short: "Remove a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			switch key {
			case "dsn":
				cfg.DSN = ""
			case "defaults.workflow":
				cfg.Defaults.Workflow = ""
			default:
				return fmt.Errorf("unknown config key: %s (supported: dsn, defaults.workflow)", key)
			}

			if err := cfg.Save(); err != nil {
				return err
			}
			cmd.Printf("Unset %s\n", key)
			return nil
		},
	}
}

func configPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}
			fmt.Println(path)
			return nil
		},
	}
}
