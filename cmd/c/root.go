package c

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

// HelpTemplate is the help template for cert-manager commands
// This uses the short and long options.
// command should not use this.
// const helpTemplate = `{{.Short}}
// Description:
//   {{.Long}}
// {{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

// UsageTemplate is the usage template for cert-manager commands
// This blocks the displaying of the global options. The main cert-manager
// command should not use this.
const usageTemplate = `Usage:{{if (and .Runnable (not .HasAvailableSubCommands))}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.UseLine}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Options:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Options:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasExample}}

Examples:
  {{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

const examplesTemplate = `
  ./{{.}} run                       Renew the internal certificates.
  ./{{.}} exec -it <container>      Renew the external certificates.
  ./{{.}} log <container>           Apply the certificates.`

const (
	rootDesc = `Container is a simple container runtime implementation.`
)

func New() *cobra.Command {
	var buf bytes.Buffer
	baseName := filepath.Base(os.Args[0])
	tmpl, _ := template.New(baseName).Parse(examplesTemplate)

	_ = tmpl.Execute(&buf, baseName)

	c := &cobra.Command{
		Use:     baseName + " [options]",
		Short:   rootDesc,
		Example: buf.String(),
		PersistentPostRun: func(cmd *cobra.Command, args []string) {},
		PreRun: func(cmd *cobra.Command, args []string) {},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add sub commands
	c.AddCommand(
		newInitCmd(),
		newRunCmd(),
		newCommitCmd(),
		newListCmd(),
		newLogsCmd(),
	)

	cobra.EnableCommandSorting = false
	c.Flags().SortFlags = false

	c.SilenceUsage = true
	c.SilenceErrors = true

	// c.DisableFlagParsing = true
	c.SetUsageTemplate(usageTemplate)
	c.DisableFlagsInUseLine = true
	return c
}
