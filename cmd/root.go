package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dbaltas/ergo/github"
	"github.com/dbaltas/ergo/jira"
	"github.com/dbaltas/ergo/repo"
	"github.com/fatih/color"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	repoURL               string
	path                  string
	skipFetch             bool
	baseBranch            string
	branchesString        string
	releaseBranchesString string
	branches              []string
	releaseBranches       []string
	organizationName      string
	repoName              string

	gc *github.Client
	jc *jira.Client
	r  *repo.Repo
)

var rootCmd = &cobra.Command{
	Use:   "ergo",
	Short: "ergo is a tool that aims to help the daily developer workflow",
	Long: `Ergo helps to
* compare multiple branches
* push to multiple branches with time interval (useful for multiple release environments)
Also it minimizing the browser interaction with github
* handles pull requests
* drafts a release
* updates release notes
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hola! type `ergo help`")
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error

		noRepoCmds := make(map[string]bool)
		noRepoCmds["help"] = true
		noRepoCmds["version"] = true

		if _, ok := noRepoCmds[cmd.Name()]; ok {
			return nil
		}

		err = initializeRepo()
		if err != nil {
			return fmt.Errorf("initialize repo: %v", err)
		}

		gc, err = github.NewClient(context.Background(), viper.GetString("github.access-token"), organizationName, repoName)
		if err != nil {
			// NOTE: ergo may still be of use without github support
			fmt.Printf("Error Initializing github %v\n", err)
		}

		tp := jira.BasicAuthTransport{
			Username: viper.GetString("jira.username"),
			Password: viper.GetString("jira.password"),
		}
		jc, err = jira.NewClient(tp.Client(), viper.GetString("jira.url"))

		if err != nil {
			fmt.Printf("Error Initializing jira client %v\n", err)
		}

		return nil
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&repoURL, "repoUrl", "", "git repo Url. ssh and https supported")
	rootCmd.PersistentFlags().StringVar(&path, "path", ".", "Location to store or retrieve from the repo")
	rootCmd.PersistentFlags().BoolVar(&skipFetch, "skipFetch", false, "Skip fetch. When set you may not be up to date with remote")

	rootCmd.PersistentFlags().StringVar(&branchesString, "branches", "", "Comma separated list of branches")
	rootCmd.PersistentFlags().StringVar(&baseBranch, "base", "", "Base branch for the comparison.")
}

func initConfig() {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	viper.AddConfigPath(home)
	viper.SetConfigName(".ergo")
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("error reading config file: %v\n", err)
		os.Exit(1)
	}
	if baseBranch == "" {
		baseBranch = viper.GetString("generic.base-branch")
	}
}

func initializeRepo() error {
	var err error
	if repoURL != "" {
		r, err = repo.NewClone(repoURL, path, viper.GetString("generic.remote"))
	} else {
		r, err = repo.NewFromPath(path, viper.GetString("generic.remote"))
	}

	if err != nil {
		return fmt.Errorf("load repo:%s", err)
	}

	if !skipFetch {
		err = r.Fetch()
		if err != nil {
			return err
		}
	}

	repoName = r.Name()
	organizationName = r.OrganizationName()

	releaseBranchesString = branchesString
	if branchesString == "" {
		branchesString = viper.GetString(fmt.Sprintf("repos.%s.status-branches", repoName))
		releaseBranchesString = viper.GetString(fmt.Sprintf("repos.%s.release-branches", repoName))
	}

	if branchesString == "" {
		branchesString = viper.GetString("generic.status-branches")
	}

	if releaseBranchesString == "" {
		releaseBranchesString = viper.GetString("generic.release-branches")
	}

	branches = strings.Split(branchesString, ",")
	releaseBranches = strings.Split(releaseBranchesString, ",")

	yellow := color.New(color.FgYellow)
	yellow.Printf("%s/%s\n", organizationName, repoName)

	return nil
}

// Execute entry point for commands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
