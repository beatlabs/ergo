package cmd

import (
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(prsCmd)
}

var prsCmd = &cobra.Command{
	Use:   "prs",
	Short: "List open pull requests [github]",
	Long:  `List open pull requests on github`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listPRs()
	},
}

func listPRs() error {
	var err error

	prs, err := gc.ListPRs()
	if err != nil {
		return err
	}

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("#", "Title", "Branch", "Url", "Creator", "Created")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, pr := range prs {
		branch := *pr.Head.Label
		if strings.HasPrefix(branch, organizationName+":") {
			branch = strings.TrimPrefix(branch, organizationName+":")
		}
		title := (*pr.Title)
		if len(title) > 60 {
			title = title[:60]
		}
		t := *pr.CreatedAt

		at := t.Format("2006-01-02 15:04")
		tbl.AddRow(*pr.Number, title, branch, *pr.HTMLURL, (*pr.User).GetLogin(), at)
	}

	tbl.Print()

	return nil
}
