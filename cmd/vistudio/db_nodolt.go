//go:build !dolt

package main

import "github.com/spf13/cobra"

func addDoltDBCommands(_ *cobra.Command) {}
