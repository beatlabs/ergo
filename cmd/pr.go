package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var compareBranch string
var title string
var description string

func init() {
	rootCmd.AddCommand(prCmd)
	prCmd.Flags().StringVar(&compareBranch, "compare", "", "The branch to compare with base branch. Defaults to current local branch.")
	prCmd.Flags().StringVar(&title, "title", "", "The title of the PR.")
	prCmd.Flags().StringVar(&description, "description", "", "The description of the PR.")
	prCmd.MarkFlagRequired("title")
}

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create a pull request [github]",
	Long:  `Create a pull request on github from compare branch to base branch`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return createPR()
	},
}

func createPR() error {
	var err error
	yellow := color.New(color.FgYellow)

	if compareBranch == "" {
		compareBranch, err = r.CurrentBranch()
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Printf(`Create a PR
	base:%s
	compare:%s
	title:%s
	description:%s
`, baseBranch, compareBranch, title, description)

	yellow.Printf("\nPress 'ok' to continue:")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	text := strings.Split(input, "\n")[0]
	if text != "ok" {
		fmt.Printf("No PR\n")
		return nil
	}

	pr, err := gc.CreatePR(baseBranch, compareBranch, title, description)

	if err != nil {
		return err
	}

	fmt.Printf("Created PR %s\n", *pr.HTMLURL)

	return nil
}
