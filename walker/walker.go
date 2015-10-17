package walker

import (
	"fmt"
	"log"

	"github.com/libgit2/git2go"
)

// UpdateRepo updates the local requirment database by parsing the git-historyz
// The string points to the location of the git database and where to create chronicle database.
func UpdateRepo(repoPath string) {
	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Repo", repo)

	head, err := repo.Head()
	if err != nil {
		panic(err)
	}
	fmt.Println("Reference, head", head.Name())

	headCommit, err := repo.LookupCommit(head.Target())
	if err != nil {
		panic(err)
	}
	fmt.Println("Head commit", headCommit.Id())

	findRoots(headCommit)

	// Find root

	// Update all line references for old req.
	// *Diff, err = repo.DiffTreeToTree(oldTree, newTree, options)

	// Check if there is some new .req files

	// Check if there is a commit req reference
	// *Diff, err = repo.DiffIndexToWorkdir(null, options)
	// Diff options https://libgit2.github.com/libgit2/#HEAD/type/git_diff_options
	// Store lines of code -> Connect to specific req.

}

func findRoots(c *git.Commit) {
	fmt.Println("Base Case", c.ParentCount())
	if c.ParentCount() == 0 {
		// base case
		// ADD commit oid
		fmt.Println("Root", c.Id())
		return
	}
	for i := uint(0); i < c.ParentCount(); i++ {
		fmt.Println("Not root", c.Id())
		findRoots(c.Parent(i))
	}
	return
}

func crawlRepo(path string) {

}
