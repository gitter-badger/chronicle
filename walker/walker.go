package walker

import (
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

	// Temporary variable used by the recursive functions
	currentCommit     git.Commit
	currentFileIsReq  bool
	currentFileBucket *bolt.Bucket
	commitReference   []string // Reference from commit msg to requirments
	isStartReqSet     bool
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

	currentTree, err := c.Tree()
	if err != nil {
		log.Fatal(err)
	}
	// Search for req files
	currentTree.Walk(indexReqFiles)

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
		blob, err := walker.currentCommit.Owner().LookupBlob(entry.Id)
		if err != nil {
			log.Fatal(err)
		}

		err = requirments.ParseReqFile(blob.Contents(), walker.db, walker.currentCommit.Author().When)
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
	walker.currentFileIsReq = walker.reqMatchString(diffDelta.NewFile.Path)
	if !walker.currentFileIsReq {
		// Create new filebucket
		// Root-bucket -> Commit-bucket (date) -> File-bucket -> (Key:Line, Value: [req]
		err := walker.db.DB.Update(func(tx *bolt.Tx) error {
			var err error
			bRoot := tx.Bucket([]byte("RootBucket"))
			bucketID := walker.currentCommit.Author().When.Format(time.RFC3339)
			bCurrent := bRoot.Bucket([]byte(bucketID))
			walker.currentFileBucket, err = bCurrent.CreateBucket([]byte(diffDelta.NewFile.Path))
			if err != nil {
				log.Fatal(err)
			}
			return err
		})
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Old file", diffDelta.OldFile)
	fmt.Println("New file", diffDelta.NewFile)
	return updateReqFromEachHunk, nil
}

func updateReqFromEachHunk(diffHunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
	// TODO: Ta bort, hjälp för alex.
	fmt.Println("Hunk Header", diffHunk.Header)
	fmt.Println("new lines", diffHunk.NewLines)
	fmt.Println("old lines", diffHunk.OldLines)
	fmt.Println("New start", diffHunk.NewStart)
	fmt.Println("Old start", diffHunk.OldStart)
	fmt.Println("")

	return updateReqFromEachLine, nil
}

func updateReqFromEachLine(diffLine git.DiffLine) error {

	if !walker.currentFileIsReq {

		// TODO: Ta bort, hjälp för alex.
		fmt.Println("New line", diffLine.NewLineno)
		fmt.Println("Old line", diffLine.OldLineno)
		fmt.Println("Num lines", diffLine.NumLines)
		fmt.Println("Content", diffLine.Content)

		switch diffLine.Origin {
		case git.DiffLineContext:
			fmt.Println("Origin Diff Line Context")
			// Line changed update old one.
		case git.DiffLineAddition:
			fmt.Println("Origin Diff Line Add")
			// If req. commit reference add to that new req.
			// Add a new start reference
			if !walker.isStartReqSet {
				// TODO: Här är jag och arbetar nu. Ska spara ner lines för nya krav.
				// err := walker.db.DB.Update(func(tx *bolt.Tx) error {
				// 	walker.currentFileBucket.Put([]byte(diffLine.NewLineno), value []byte)
				// 	if err != nil {
				// 		log.Fatal(err)
				// 	}
				// 	return err
				// })
			}
		case git.DiffLineDeletion:
			fmt.Println("Origin Diff Line delete")
			// Decrese req. which have this line
		case git.DiffLineContextEOFNL:
			log.Fatal("GIT_DIFF_LINE_CONTEXT_EOFNL")
		case git.DiffLineAddEOFNL:
			fmt.Println("Origin GIT_DIFF_LINE_ADD_EOFNL")

			// OUTPUT FROM TO LAST LINE:
			// New line -1
			// Old line 29
			// Num lines 2
			// Origin GIT_DIFF_LINE_ADD_EOFNL
			// Content
			// \ No newline at end of file
		case git.DiffLineDelEOFNL:
			fmt.Println("Origin Diff Line delete end of file NL")
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
		fmt.Println("")
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
