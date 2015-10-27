package walker

import (
	"crypto/rand"
	"fmt"
	"log"

	"github.com/libgit2/git2go"
)

// RandString generates a random alphanumerical string in desired length
func RandString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

// Stash is a replication of the stash feature in git but do create real branches.
// Existing branches will be overwritten
func Stash(branchName string, repo *git.Repository) error {
	fmt.Println("Stashing files to branch:", branchName)
	head, err := repo.Head()
	if err != nil {
		log.Fatal(err)
	}
	headCommit, err := repo.LookupCommit(head.Target())
	if err != nil {
		log.Fatal(err)
	}
	// True means that existing branch will be overwritten. Use with care.
	branch, err := repo.CreateBranch(branchName, headCommit, true)
	if err != nil {
		log.Fatal(err)
	}
	// Get the index also called the staging area.
	idx, err := repo.Index()
	if err != nil {
		log.Fatal(err)
	}
	// Writes the current states of all files to the index (staging).
	// TODO: Check if this needs to be written as "./*"
	root := []string{"*"}
	err = idx.UpdateAll(root, UpdateIndexStashCallback)
	if err != nil {
		log.Fatal(err)
	}

	// This generate a tree from current index.
	// https://libgit2.github.com/libgit2/#HEAD/group/index/git_index_write_tree
	treeID, err := idx.WriteTree()
	if err != nil {
		log.Fatal(err)
	}

	// This will write the new modified index to the git database
	err = idx.Write()
	if err != nil {
		panic(err)
	}

	// Get the generated tree represantation of the above id
	tree, err := repo.LookupTree(treeID)
	if err != nil {
		log.Fatal(err)
	}
	// Get the parent commit in this case the branch
	commitTarget, err := repo.LookupCommit(branch.Target())
	if err != nil {
		log.Fatal(err)
	}
	// Commit the tree as child to branch commit (commitTarget).
	// Update the /refs/heads so we can find back to this commit by only knowing repo name.
	message := "Stashing commit by chronicle"
	_, err = repo.CreateCommit("refs/heads/"+branchName, walker.commitSignature, walker.commitSignature, message, tree, commitTarget)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

// UnStash moves content of a branch back to index, staging area.
// Existing branches will be overwritten
func UnStash(branchName string, repo *git.Repository) error {
	// See if branch exist
	branch, err := repo.LookupBranch(branchName, git.BranchLocal)
	if err != nil {
		log.Fatal(err)
	}
	// Temporary save the current head
	currentReference, err := repo.Head()
	if err != nil {
		log.Fatal(err)
	}
	// Set the stashed branch as head
	repo.SetHead("refs/heads/" + branchName)
	if err != nil {
		log.Fatal(err)
	}
	// Check out and write over the staging area.
	err = repo.CheckoutHead(walker.checkoutOptions)
	if err != nil {
		log.Fatal(err)
	}
	// Reset the head to one before
	repo.SetHead(currentReference.Name())
	if err != nil {
		log.Fatal(err)
	}
	// Delete the old Stash-branch
	err = branch.Delete()
	if err != nil {
		log.Fatal(err)
	}
	return err
}

// UpdateIndexStashCallback always return 0 and thereby always updates index files.
func UpdateIndexStashCallback(path string, pathspec string) int {
	return 0
}
