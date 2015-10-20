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
	diffOptions     *git.DiffOptions
	diffDetail      git.DiffDetail
	currentCommit   git.Commit
	db              *database.Database
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
	diffOpt, _ := git.DefaultDiffOptions()
	walker.diffOptions = &diffOpt
	// Set resolution of diffs, 0 = file, 1 = Hunk, 2 = line by line
	walker.diffDetail = git.DiffDetailLines
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
	crawlRepo(headCommit)

	// Check if there is a commit req reference
	// *Diff, err = repo.DiffIndexToWorkdir(null, options)
	// Diff options https://libgit2.github.com/libgit2/#HEAD/type/git_diff_options
	// Store lines of code -> Connect to specific req.
}

func crawlRepo(c *git.Commit) error {
	walker.currentCommit = *c
	walker.db.DB.Update(func(tx *bolt.Tx) error {
		bRoot := tx.Bucket([]byte("RootBucket"))
		fmt.Println("Zero", c.Author().When.Format(time.RFC3339))
		bCurrentTime, err := bRoot.CreateBucketIfNotExists([]byte(c.Author().When.Format(time.RFC3339)))
		if err != nil {
			panic(err)
		}
		bCurrentTime.Put([]byte("commit"), []byte(c.Id().String()))
		return err
	})

	if c.ParentCount() == 0 {
		// base case
		fmt.Println("Root", c.Id())
		return nil
	}
	for i := uint(0); i < c.ParentCount(); i++ {
		fmt.Println("Not root", c.Id())
		crawlRepo(c.Parent(i))
		walker.currentCommit = *c
	}

	fmt.Println("====== CLIMBING UP THE TREE =======   ", walker.currentCommit.Id())
	// Search for .req files
	currentTree, err := c.Tree()
	if err != nil {
		log.Fatal(err)
	}
	currentTree.Walk(indexReqFiles)

	// Check if there is a commit reference.
	commitReferences()

	// Check if req->code have changed
	// Check with parents commits tree, eg. c.Parent(i).Tree. <- OLD one c.Tree <- NEW one
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

	// fmt.Println("Path to file:", s)
	// fmt.Println("Is .req file:", walker.reqMatchString(entry.Name))
	// fmt.Println("Name", entry.Name)
	// fmt.Println("Filemode", entry.Filemode)
	// fmt.Println("Type", entry.Type)
	// fmt.Println("")
	if walker.reqMatchString(entry.Name) {
		fmt.Println("First", walker.currentCommit.Author().When.Format(time.RFC3339))
		requirments.ParseReqFile("./"+s+entry.Name, walker.db, walker.currentCommit.Author().When)
	}
	return 0
}

func updateReqFromEachFile(diffDelta git.DiffDelta, nbr float64) (git.DiffForEachHunkCallback, error) {
	fmt.Println("EachFile", nbr)
	return updateReqFromEachHunk, nil
}

func updateReqFromEachHunk(diffHunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
	// fmt.Println("EachHunk", diffHunk.Header)
	// fmt.Println("NewLines", diffHunk.NewLines)
	// fmt.Println("NewStart", diffHunk.NewStart)
	// fmt.Println("OldLines", diffHunk.OldLines)
	// fmt.Println("OldStart", diffHunk.OldStart)
	return updateReqFromEachLine, nil
}

func updateReqFromEachLine(diffLine git.DiffLine) error {
	// fmt.Println("New line", diffLine.NewLineno)
	// fmt.Println("Old line", diffLine.OldLineno)
	// fmt.Println("Num lines", diffLine.NumLines)
	// fmt.Println("Origin", diffLineToString(diffLine.Origin))
	// fmt.Println("Content", diffLine.Content)
	// fmt.Println("")
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
		log.Fatal("GIT_DIFF_LINE_ADD_EOFNL")
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

	fmt.Println("Printing commit msg:", msg)
	for _, ref := range reqs {
		if utf8.RuneCountInString(ref) == 9 {
			fmt.Println("All code references to commit", ref)
		} else {
			fmt.Println("Store specific commit. TODO", ref)
		}
	}
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

func diffLineToString(i git.DiffLineType) string {
	switch i {
	case git.DiffLineContext:
		return "GIT_DIFF_LINE_CONTEXT"
	case git.DiffLineAddition:
		return "GIT_DIFF_LINE_ADDITION"
	case git.DiffLineDeletion:
		return "GIT_DIFF_LINE_DELETION"
	case git.DiffLineContextEOFNL:
		return "GIT_DIFF_LINE_CONTEXT_EOFNL"
	case git.DiffLineAddEOFNL:
		return "GIT_DIFF_LINE_ADD_EOFNL"
	case git.DiffLineDelEOFNL:
		return "GIT_DIFF_LINE_DEL_EOFNL"
	case git.DiffLineFileHdr:
		return "GIT_DIFF_LINE_FILE_HDR"
	case git.DiffLineHunkHdr:
		return "GIT_DIFF_LINE_HUNK_HDR"
	case git.DiffLineBinary:
		return "GIT_DIFF_LINE_BINARY"
	}
	return "Unknown"
}
