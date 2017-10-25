package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/client"
	"github.com/spf13/cobra"
)

// table output settings
const (
	minWidth = 0
	tabWidth = 8
	padding  = 4
	padChar  = ' '
)

var appSlugOrID string
var apiToken string
var apiOrigin = "https://api.replicated.com/vendor"

func init() {
	RootCmd.PersistentFlags().StringVar(&appSlugOrID, "app", "", "The app slug or app id to use in all calls")
	RootCmd.PersistentFlags().StringVar(&apiToken, "token", "", "The API token to use to access your app in the Vendor API")

	originFromEnv := os.Getenv("REPLICATED_API_ORIGIN")
	if originFromEnv != "" {
		apiOrigin = originFromEnv
	}
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "replicated",
	Short: "Manage channels and releases",
	Long:  `The replicated CLI allows vendors to manage their apps' channels and releases.`,
}

// Almost the same as the default but don't print help subcommand
var rootCmdUsageTmpl = `
Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

  Examples:
  {{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
  {{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
  {{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	fmt.Println("-- start --")
	w := tabwriter.NewWriter(stdout, minWidth, tabWidth, padding, padChar, tabwriter.TabIndent)

	// get api client and app ID after flags are parsed
	runCmds := &runners{
		stdin: stdin,
		w:     w,
	}
	if stderr != nil {
		RootCmd.SetOutput(stderr)
	}

	channelCreateCmd.RunE = runCmds.channelCreate
	channelInspectCmd.RunE = runCmds.channelInspect
	channelAdoptionCmd.RunE = runCmds.channelAdoption
	channelReleasesCmd.RunE = runCmds.channelReleases
	channelCountsCmd.RunE = runCmds.channelCounts
	channelLsCmd.RunE = runCmds.channelList
	channelRmCmd.RunE = runCmds.channelRemove
	releaseCreateCmd.RunE = runCmds.releaseCreate
	releaseInspectCmd.RunE = runCmds.releaseInspect
	releaseLsCmd.RunE = runCmds.releaseList
	releaseUpdateCmd.RunE = runCmds.releaseUpdate
	releasePromoteCmd.RunE = runCmds.releasePromote

	RootCmd.SetUsageTemplate(rootCmdUsageTmpl)

	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		log.Println("-- here --")
		if apiToken == "" {
			apiToken = os.Getenv("REPLICATED_API_TOKEN")
			if apiToken == "" {
				return errors.New("Please provide your API token")
			}
		}
		api := client.NewHTTPClient(apiOrigin, apiToken)
		runCmds.api = api

		if appSlugOrID == "" {
			appSlugOrID = os.Getenv("REPLICATED_APP")
			if appSlugOrID == "" {
				return errors.New("App not provided")
			}
		}

		fmt.Printf("-- appSlugOrID --\n %v\n", appSlugOrID)

		app, err := api.GetApp(appSlugOrID)
		if err != nil {
			return err
		}
		runCmds.appID = app.Id

		return nil
	}

	return RootCmd.Execute()
}
