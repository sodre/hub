package commands

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/github/hub/github"
	"github.com/github/hub/ui"
	"github.com/github/hub/utils"
)

var cmdDelete = &Command{
	Run:   delete,
	Usage: "delete [-y] [<ORGANIZATION>/]<NAME>",
	Long: `Delete an existing repository on GitHub.

## Options:

	-y, --yes
		Skip the confirmation prompt and immediately delete the repository.

	[<ORGANIZATION>/]<NAME>
		The name for the repository on GitHub.

## Examples:
		$ hub delete recipes
		[ personal repo deleted on GitHub ]

		$ hub delete sinatra/recipes
		[ repo deleted in GitHub organization ]

## See also:

hub-init(1), hub(1)
`,
}

var (
	flagDeleteAssumeYes bool
)

func init() {
	cmdDelete.Flag.BoolVarP(&flagDeleteAssumeYes, "--yes", "y", false, "YES")

	CmdRunner.Use(cmdDelete)
}

func delete(command *Command, args *Args) {
	var repoName string
	if args.IsParamsEmpty() {
		ui.Errorln("Expecting name of repository to delete")
		ui.Errorln(command.Synopsis())
		os.Exit( 1 );
	} else {
		reg := regexp.MustCompile("^[^-]")
		if !reg.MatchString(args.FirstParam()) {
			err := fmt.Errorf("invalid argument: %s", args.FirstParam())
			utils.Check(err)
		}
		repoName = args.FirstParam()
	}

	config := github.CurrentConfig()
	host, err := config.DefaultHost()
	if err != nil {
		utils.Check(github.FormatError("deleting repository", err))
	}

	owner := host.User
	if strings.Contains(repoName, "/") {
		split := strings.SplitN(repoName, "/", 2)
		owner = split[0]
		repoName = split[1]
	}

	project := github.NewProject(owner, repoName, host.Host)
	gh := github.NewClient(project.Host)
	
	if !flagDeleteAssumeYes {
		ui.Printf("Really delete repository '%s' (yes/N)? ", project)
		answer := ""
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			answer = strings.TrimSpace(scanner.Text())
		}
		utils.Check(scanner.Err())
		if answer != "yes" {
			utils.Check(fmt.Errorf("Please type 'yes' for confirmation."))
		}
	}

	err = gh.DeleteRepository(project)
	if err != nil && strings.Contains(err.Error(), "HTTP 403") {
		fmt.Println("Please edit the token used for hub at https://github.com/settings/tokens\nand verify that the `delete_repo` scope is enabled.\n")
	}
	utils.Check(err)

	fmt.Printf("Deleted repository %v\n", repoName)

	args.NoForward()
}
