package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var compareBranch string
var title string
var description string
var number int
var reviewers string
var teamReviewers string

func init() {
	rootCmd.AddCommand(prCmd)
	prCmd.Flags().StringVar(&compareBranch, "compare", "", "The branch to compare with base branch. Defaults to current local branch.")
	prCmd.Flags().StringVar(&title, "title", "", "The title of the PR.")
	prCmd.Flags().StringVar(&reviewers, "reviewers", "", "Add reviewers.")
	prCmd.Flags().StringVar(&teamReviewers, "teamReviewers", "", "Add a team as reviewers.")
	prCmd.Flags().StringVar(&description, "description", "", "The description of the PR.")
}

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create or show a pull request [github]",
	Long: `Create a pull request on github from compare branch to base branch
	ergo pr --title "my new pull request --reviewers pespantelis,nstratos,mantzas"
Show details of a pr by pr number
	ergo pr 18
Add reviewers to an existing pr
	ergo pr 18 --reviewers pespantelis,nstratos,mantzas"
	`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// show a PR by number
		if len(args) > 0 {
			number, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			err = getPR(number)
			if err != nil {
				return err
			}

			if reviewers != "" || teamReviewers != "" {
				_, err = gc.RequestReviewersForPR(number, reviewers, teamReviewers)
				if err != nil {
					return err
				}
			}

			return nil
		}

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

	if reviewers != "" || teamReviewers != "" {
		_, err = gc.RequestReviewersForPR(pr.GetNumber(), reviewers, teamReviewers)
		if err != nil {
			return err
		}
	}

	return nil
}

func getPR(number int) error {
	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	prp, err := gc.GetPR(number)
	if err != nil {
		return err
	}
	pr := *prp

	fmt.Println()
	green.Printf("#%d: %s\n", pr.GetNumber(), pr.GetTitle())
	fmt.Printf("into:%s from:%s\n", yellow.Sprint(pr.Base.GetLabel()), yellow.Sprint(pr.Head.GetLabel()))
	if pr.GetBody() != "" {
		yellow.Println(pr.GetBody())
	}
	fmt.Println()

	fmt.Printf("%s: %d\n", yellow.Sprint("# Commits"), pr.GetCommits())
	fmt.Printf("%s:%s, %s:%s, %s:%s\n",
		yellow.Sprint("created"), pr.GetCreatedAt().Format("2006-01-02 15:04"),
		yellow.Sprint("modified"), pr.GetUpdatedAt().Format("2006-01-02 15:04"),
		yellow.Sprint("merged"), pr.GetMergedAt().Format("2006-01-02 15:04"),
	)
	fmt.Println()

	a := green.Sprintf("%d", pr.GetAdditions())
	d := red.Sprintf("%d", pr.GetDeletions())
	c := yellow.Sprintf("%d", pr.GetChangedFiles())
	fmt.Printf("%s files changed, %s additions, %s deletions\n", c, a, d)

	return nil
}
