package walker

import (
	"fmt"
	"log"
	"regexp"

	"github.com/libgit2/git2go"
)

var walker Walker

// Walker is a container for storing results and holding working struct like regex matchers.
type Walker struct {
	reqMatcher *regexp.Regexp
}

// Wraper for regex matcher for .req file
func (w *Walker) reqMatchString(s string) bool {
	return w.reqMatcher.MatchString(s)
}

// UpdateRepo updates the local requirment database by parsing the git-historyz
// The string points to the location of the git database and where to create chronicle database.
func UpdateRepo(repoPath string) {
	walker = Walker{}
	walker.reqMatcher, _ = regexp.Compile(".*\\.req")

	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Repo", repo)

	head, err := repo.Head()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Reference, head", head.Name())

	headCommit, err := repo.LookupCommit(head.Target())
	if err != nil {
		panic(err)
	}
	fmt.Println("Head commit", headCommit.Id())

	rootIds := crawlRepo(headCommit)
	fmt.Println("All roots of git tree", rootIds)

	// Update all line references for old req.
	// *Diff, err = repo.DiffTreeToTree(oldTree, newTree, options)

	// Check if there is a commit req reference
	// *Diff, err = repo.DiffIndexToWorkdir(null, options)
	// Diff options https://libgit2.github.com/libgit2/#HEAD/type/git_diff_options
	// Store lines of code -> Connect to specific req.

}

func crawlRepo(c *git.Commit) []*git.Oid {
	var rootOid []*git.Oid
	if c.ParentCount() == 0 {
		// base case
		fmt.Println("Root", c.Id())
		rootOid = []*git.Oid{c.Id()}
		return rootOid
	}
	for i := uint(0); i < c.ParentCount(); i++ {
		fmt.Println("Not root", c.Id())
		rootOid = append(rootOid, crawlRepo(c.Parent(i))...)
	}
	// Search for .req files
	// -> Update database, new and removed req.
	currentTree, err := c.Tree()
	if err != nil {
		log.Fatal(err)
	}
	currentTree.Walk(searchForReqFiles)

	// Check if req->code have changed
	// Check with parents commits tree, eg. c.Parent(i).Tree. <- OLD one c.Tree <- NEW one

	// Check if there is a commit reference.

	return rootOid
}

func searchForReqFiles(s string, entry *git.TreeEntry) int {

	fmt.Println("Path to file:", s)
	fmt.Println("Is .req file:", walker.reqMatchString(entry.Name))
	fmt.Println("Name", entry.Name)
	fmt.Println("Filemode", entry.Filemode)
	fmt.Println("Type", entry.Type)
	fmt.Println("")
	return 0
}
