package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/yp/internal/clip"
	"github.com/jamesonstone/yp/internal/picker"
)

var noHidden bool

func Execute() error {
	return newRootCmd(os.Stdout).Execute()
}

func newRootCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yp [path]",
		Short: "Yank an absolute path to the clipboard",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			start := "."
			if len(args) > 0 {
				start = args[0]
			}

			resolved, err := filepath.Abs(start)
			if err != nil {
				return err
			}
			resolved = filepath.Clean(resolved)

			target := resolved
			if info, err := os.Stat(resolved); err == nil && info.IsDir() {
				target, err = picker.Run(resolved, !noHidden)
				if err != nil {
					return err
				}
			}

			target = filepath.Clean(target)
			if err := clip.Write(target); err != nil {
				return err
			}

			_, err = fmt.Fprintf(out, "📋 %s\n", target)
			return err
		},
	}

	cmd.Flags().BoolVar(&noHidden, "no-hidden", false, "exclude hidden files")
	cmd.SetOut(out)
	cmd.SetErr(os.Stderr)
	return cmd
}
