package walker

import (
	"crypto/rand"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/Benefactory/chronicle/database"
	"github.com/Benefactory/chronicle/requirments"
	"github.com/boltdb/bolt"
	"github.com/libgit2/git2go"
)

var walker Walker

// Walker is a container for storing results and holding working struct like regex matchers.
type Walker struct {
	reqMatcher      *regexp.Regexp
	commitMatcher   *regexp.Regexp
	db              *database.Database
	diffOptions     *git.DiffOptions
	checkoutOptions *git.CheckoutOpts
	diffDetail      git.DiffDetail
	commitSignature *git.Signature

	// Temporary variable used by the recursive functions
	currentCommit   git.Commit
	currentIsReq    bool
	commitReference []string
}

// Wraper for regex matcher for .req file
func (w *Walker) reqMatchString(s string) bool {
	return w.reqMatcher.MatchString(s)
}

// UpdateRepo updates the local requirment database by parsing the git-history.
// The string points to the location of the git database and where to create chronicle database.
func UpdateRepo(rootPath string, db *database.Database) {
	walker = Walker{}
	walker.reqMatcher, _ = regexp.Compile(".*\\.req")
	walker.commitMatcher, _ = regexp.Compile("#[a-zA-Z0-9]{8}(:?( \\([0-9a-zA-Z_, \\-.]*\\))*)")
	// Diff options https://libgit2.github.com/libgit2/#HEAD/type/git_diff_options
	diffOpt, _ := git.DefaultDiffOptions()
	walker.diffOptions = &diffOpt
	walker.checkoutOptions = &git.CheckoutOpts{
		Strategy: git.CheckoutForce,
	}
	// Set resolution of diffs, 0 = file, 1 = Hunk, 2 = line by line
	walker.diffDetail = git.DiffDetailLines
	// Create signature for the walker.
	walker.commitSignature = &git.Signature{
		Name:  "Chronicle tool",
		Email: "chronicle@benefactory.se",
		When:  time.Now(),
	}
	walker.db = db

	db.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("RootBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	repo, err := git.OpenRepository("." + string(filepath.Separator) + rootPath)
	if err != nil {
		log.Fatal(err)
	}
	head, err := repo.Head()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Current branch:", head.Name())

	headCommit, err := repo.LookupCommit(head.Target())
	if err != nil {
		panic(err)
	}
	fmt.Println("Head commit", headCommit.Id())
	fmt.Println("")
	crawlRepo(headCommit)
}

func crawlRepo(c *git.Commit) error {
	walker.currentCommit = *c
	// Create a boltdb bucket for each commit. Use time for key to enable sorting.
	walker.db.DB.Update(func(tx *bolt.Tx) error {
		bRoot := tx.Bucket([]byte("RootBucket"))
		bCurrentTime, err := bRoot.CreateBucketIfNotExists([]byte(c.Author().When.Format(time.RFC3339)))
		if err != nil {
			panic(err)
		}
		bCurrentTime.Put([]byte("commit"), []byte(c.Id().String()))
		return err
	})

	// BASE CASE
	if c.ParentCount() == 0 {
		fmt.Println("Root:", c.Id())
		fmt.Println("")
		return nil
	}
	// DIG DEEPER
	for i := uint(0); i < c.ParentCount(); i++ {
		fmt.Println("Digging deeper:", c.Id())
		fmt.Println("")
		crawlRepo(c.Parent(i))
		walker.currentCommit = *c
	}

	fmt.Println("")
	fmt.Println("========= CLIMBING UP THE TREE =========")
	fmt.Println("Current commit:", walker.currentCommit.Id())
	// Search for .req files
	currentTree, err := c.Tree()
	if err != nil {
		log.Fatal(err)
	}
	err = Stash("ChronicleStash", c.Owner())
	if err != nil {
		log.Fatal(err)
	}
	// Unlike git checkout, it does not move the HEAD commit for you.
	// https://libgit2.github.com/libgit2/#HEAD/type/git_checkout_strategy_t
	err = c.Owner().CheckoutTree(currentTree, walker.checkoutOptions)
	if err != nil {
		log.Fatal(err)
	}
	currentTree.Walk(indexReqFiles)

	err = UnStash("ChronicleStash", c.Owner())
	if err != nil {
		log.Fatal(err)
	}
	// Check if there is a commit reference.
	commitReferences()
	// Create a diff between current and parent tree
	for i := uint(0); i < c.ParentCount(); i++ {
		parrentTree, err := c.Parent(i).Tree()
		if err != nil {
			log.Fatal(err)
		}
		diff, err := c.Owner().DiffTreeToTree(parrentTree, currentTree, walker.diffOptions)
		if err != nil {
			log.Fatal(err)
		}
		diff.ForEach(updateReqFromEachFile, walker.diffDetail)
	}
	return nil
}

func indexReqFiles(s string, entry *git.TreeEntry) int {
	if walker.reqMatchString(entry.Name) {
		err := requirments.ParseReqFile("./"+s+entry.Name, walker.db, walker.currentCommit.Author().When)
		if err != nil {
			fmt.Println("")
			fmt.Println("TOML FORMAT ERROR")
			fmt.Println("File:\"", s+entry.Name, "\" ignored")
			fmt.Println(err)
			fmt.Println("")
		}
	}
	return 0
}

func updateReqFromEachFile(diffDelta git.DiffDelta, nbr float64) (git.DiffForEachHunkCallback, error) {
	walker.currentIsReq = walker.reqMatchString(diffDelta.NewFile.Path)
	fmt.Println("Old file", diffDelta.OldFile)
	fmt.Println("New file", diffDelta.NewFile)
	return updateReqFromEachHunk, nil
}

func updateReqFromEachHunk(diffHunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
	return updateReqFromEachLine, nil
}

func updateReqFromEachLine(diffLine git.DiffLine) error {

	switch diffLine.Origin {
	case git.DiffLineContext:
		// Line changed update old one.
	case git.DiffLineAddition:
		// If req. commit reference add to that new req.
	case git.DiffLineDeletion:
		// Decrese req. which have this line
	case git.DiffLineContextEOFNL:
		log.Fatal("GIT_DIFF_LINE_CONTEXT_EOFNL")
	case git.DiffLineAddEOFNL:
		fmt.Println("New line", diffLine.NewLineno)
		fmt.Println("Old line", diffLine.OldLineno)
		fmt.Println("Num lines", diffLine.NumLines)
		fmt.Println("Origin GIT_DIFF_LINE_ADD_EOFNL")
		fmt.Println("Content", diffLine.Content)
		fmt.Println("")

		// OUTPUT FROM TO LAST LINE:
		// New line -1
		// Old line 29
		// Num lines 2
		// Origin GIT_DIFF_LINE_ADD_EOFNL
		// Content
		// \ No newline at end of file
	case git.DiffLineDelEOFNL:
		// Line is deleted at the end of a file? Update req. by decresing the line count?
		//
		// OUTPUT FROM TO LAST LINES:
		// New line 29
		// Old line -1
		// Num lines 0
		// Origin GIT_DIFF_LINE_ADDITION
		// Content .chronicle
		//
		// New line 29
		// Old line -1
		// Num lines 2
		// Origin GIT_DIFF_LINE_DEL_EOFNL
		// Content
		// \ No newline at end of file
	case git.DiffLineFileHdr:
		log.Fatal("GIT_DIFF_LINE_FILE_HDR")
	case git.DiffLineHunkHdr:
		log.Fatal("GIT_DIFF_LINE_HUNK_HDR")
	case git.DiffLineBinary:
		log.Fatal("GIT_DIFF_LINE_BINARY")
	}
	return nil
}

func commitReferences() {
	msg := walker.currentCommit.Message()
	reqs := walker.commitMatcher.FindStringSubmatch(msg)

	fmt.Println("Commit message:", msg)
	for _, ref := range reqs {
		if utf8.RuneCountInString(ref) == 9 {
			// Currently only support one ref all. fix.
			fmt.Println("Code references to commit:", ref)
			walker.commitReference = append(walker.commitReference, ref)
		} else {
			fmt.Println("Specific code refere to commit (NOT IMPLEMENTED)", ref)
		}
	}
	fmt.Println("")
}

func randString(n int) string {
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
	// This created index can later be commited.
	// https://libgit2.github.com/libgit2/#HEAD/group/index/git_index_write_tree
	treeID, err := idx.WriteTree()
	if err != nil {
		log.Fatal(err)
	}
	// Get tree represantation of the above id
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

	// 	# ... hack hack hack ...
	// $ git checkout -b my_wip
	// $ git commit -a -m "WIP"
	// $ git checkout master
	// $ edit emergency fix
	// $ git commit -a -m "Fix in a hurry"
	// $ git checkout my_wip
	// $ git reset --soft HEAD^
	// # ... continue hacking ...
}

// UnStash moves content of a branch back to index, staging area.
// Existing branches will be overwritten
func UnStash(branchName string, repo *git.Repository) error {
	referenceNameIterator, err := repo.NewReferenceNameIterator()
	if err != nil {
		log.Fatal(err)
	}
	for err == nil {
		var referenceString string
		var reference *git.Reference
		referenceString, err = referenceNameIterator.Next()
		reference, err = referenceNameIterator.ReferenceIterator.Next()
		fmt.Println("Reference string:", referenceString)
		fmt.Println("Reference:", reference)
	}

	return err
}
