package walker

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"unicode/utf8"

	"github.com/Benefactory/chronicle/database"
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
	rootBucket               *bolt.Bucket
	currentCommit            git.Commit
	currentFileIsReq         bool
	currentFileReference     database.FileReferences
	parentFileReference      database.FileReferences
	currentAddLine           int
	currentCommitBucket      *bolt.Bucket
	parentCommitBucket       *bolt.Bucket
	copiedFileReferences     map[string]bool
	parentFileReferenceValid bool
	commitReference          []string // Reference from commit msg to requirments
	isFileSaved              bool     // Used to prevent saving data twice
	isStartReqSet            bool
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

	walker.isFileSaved = true // Prevent saving nil
	crawlRepo(headCommit)
}

func crawlRepo(c *git.Commit) error {
	walker.currentCommit = *c
	// Create a boltdb bucket for each commit.
	walker.db.DB.Update(func(tx *bolt.Tx) error {
		bRoot := tx.Bucket([]byte("RootBucket"))
		_, err := bRoot.CreateBucketIfNotExists([]byte(c.Id().String()))
		if err != nil {
			panic(err)
		}
		err = bRoot.Put([]byte(c.Author().When.String()), []byte(c.Id().String()))
		if err != nil {
			panic(err)
		}
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

		// Open bucket to copy last commits refs
		walker.db.DB.Update(func(tx *bolt.Tx) error {
			walker.copiedFileReferences = make(map[string]bool)
			walker.rootBucket = tx.Bucket([]byte("RootBucket"))
			walker.parentCommitBucket = walker.rootBucket.Bucket([]byte(c.Parent(i).Id().String()))
			walker.currentCommitBucket = walker.rootBucket.Bucket([]byte(c.Id().String()))
			diff, err := c.Owner().DiffTreeToTree(parrentTree, currentTree, walker.diffOptions)
			if err != nil {
				log.Fatal(err)
			}
			diff.ForEach(updateReqFromEachFile, walker.diffDetail)

			// Copy FileReferences from parent CommitBucket to current CommitBucket.
			// Exclude all files already handled in diffs
			c := walker.parentCommitBucket.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				oid := string(k[:])
				if _, ok := walker.copiedFileReferences[oid]; !ok {
					// Copy fileReference
					walker.currentCommitBucket.Put(k, v)
				}
			}
			return nil
		})
	}

	return nil
}

func indexReqFiles(s string, entry *git.TreeEntry) int {
	if walker.reqMatchString(entry.Name) {
		_, err := walker.currentCommit.Owner().LookupBlob(entry.Id)
		if err != nil {
			log.Fatal(err)
		}

		// err = requirments.ParseReqFile(blob.Contents(), walker.db, walker.currentCommit.Author().When)
		// if err != nil {
		// 	fmt.Println("")
		// 	fmt.Println("TOML FORMAT ERROR")
		// 	fmt.Println("File:\"", s+entry.Name, "\" ignored")
		// 	fmt.Println(err)
		// 	fmt.Println("")
		// }
	}
	return 0
}

func updateReqFromEachFile(diffDelta git.DiffDelta, nbr float64) (git.DiffForEachHunkCallback, error) {
	// Add this file to copy ignore list
	walker.copiedFileReferences[diffDelta.OldFile.Oid.String()] = true
	// Last New LineRequirmentEnd from previous file
	if walker.isStartReqSet {
		addLineEnd()
	}
	// Save last files references to boltDB
	if !walker.isFileSaved {
		saveFileReference()
	}

	if walker.parentCommitBucket != nil { // Root commits have no parents

		byteData := walker.parentCommitBucket.Get([]byte(diffDelta.OldFile.Oid.String()))
		fmt.Println("Parent commit bucket exist")
		fmt.Println("ByteData", byteData)
		if byteData != nil {
			buffer := bytes.NewBuffer(byteData)
			dec := gob.NewDecoder(buffer)
			err := dec.Decode(&walker.parentFileReference)
			if err != nil {
				log.Fatal("decode:", err)
			}
			fmt.Println("Decode success", walker.parentFileReference.FileID)
			walker.parentFileReferenceValid = true
		} else {
			walker.parentFileReferenceValid = false
		}
	}

	walker.currentAddLine = 0 // Reset add line count for every file
	walker.currentFileIsReq = walker.reqMatchString(diffDelta.NewFile.Path)
	if !walker.currentFileIsReq {
		lineReferences := make(map[int][]database.Reference)
		walker.currentFileReference = database.FileReferences{*diffDelta.NewFile.Oid, lineReferences}
		walker.isFileSaved = false
	}

	fmt.Println("Old file", diffDelta.OldFile)
	fmt.Println("New file", diffDelta.NewFile)
	return updateReqFromEachHunk, nil
}

func updateReqFromEachHunk(diffHunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
	return updateReqFromEachLine, nil
}

func updateReqFromEachLine(diffLine git.DiffLine) error {

	if !walker.currentFileIsReq {
		switch diffLine.Origin {
		case git.DiffLineContext:
			// New LineRequirmentEnd
			if walker.isStartReqSet {
				addLineEnd()
			}

			// fmt.Println("Origin Diff Line Context ", diffLine.NewLineno)
			// fmt.Println(diffLine.OldLineno)
			// Line changed update old one.
		case git.DiffLineAddition:
			if walker.currentAddLine != diffLine.NewLineno {
				// New LineRequirmentEnd
				if walker.isStartReqSet {
					addLineEnd()
				}
				// New LineRequirmentStart
				addLineStart(diffLine.NewLineno)
				// Update line tracking for the new block
				walker.currentAddLine = diffLine.NewLineno + 1
			} else {
				// Increase add block count
				walker.currentAddLine++
			}

		case git.DiffLineDeletion:
			if walker.isStartReqSet {
				addLineEnd()
			}
			// fmt.Println("Origin Diff Line delete")
			// Decrese req. which have this line
		case git.DiffLineContextEOFNL:
			if walker.isStartReqSet {
				addLineEnd()
			}
			log.Fatal("GIT_DIFF_LINE_CONTEXT_EOFNL")
		case git.DiffLineAddEOFNL:
			if walker.isStartReqSet {
				addLineEnd()
			}
			// fmt.Println("Origin GIT_DIFF_LINE_ADD_EOFNL")

			// OUTPUT FROM TO LAST LINE:
			// New line -1
			// Old line 29
			// Num lines 2
			// Origin GIT_DIFF_LINE_ADD_EOFNL
			// Content
			// \ No newline at end of file
		case git.DiffLineDelEOFNL:
			if walker.isStartReqSet {
				addLineEnd()
			}
			// fmt.Println("Origin Diff Line delete end of file NL")
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
			if walker.isStartReqSet {
				addLineEnd()
			}
			log.Fatal("GIT_DIFF_LINE_FILE_HDR")
		case git.DiffLineHunkHdr:
			if walker.isStartReqSet {
				addLineEnd()
			}
			log.Fatal("GIT_DIFF_LINE_HUNK_HDR")
		case git.DiffLineBinary:
			if walker.isStartReqSet {
				addLineEnd()
			}
			log.Fatal("GIT_DIFF_LINE_BINARY")
		}
	}
	return nil
}

func commitReferences() {
	msg := walker.currentCommit.Message()
	reqs := walker.commitMatcher.FindStringSubmatch(msg)
	walker.commitReference = []string{}
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

func saveFileReference() error {

	var data bytes.Buffer // foo := bufio.NewWriter(&data)?
	enc := gob.NewEncoder(&data)
	err := enc.Encode(walker.currentFileReference)
	if err != nil {
		log.Fatal("encode:", err)
		fmt.Println("Encode Error")
	}
	err = walker.currentCommitBucket.Put([]byte(walker.currentFileReference.FileID.String()), data.Bytes())
	if err != nil {
		log.Fatal("Put error:", err)
	}

	fmt.Println("Saved fileReferences with id:", walker.currentFileReference.FileID.String())
	if err != nil {
		log.Fatal(err)
	}
	// Prevent the reference to be saved twice
	walker.isFileSaved = true
	return err
}

func addLineStart(line int) {
	// New LineRequirmentStart
	var references []database.Reference
	for _, ref := range walker.commitReference {
		references = append(references, database.Reference{ref, database.LineRequirmentStart})
		fmt.Println("Adding LineRequirmentStart: [reference, line]", ref, line)
	}
	if references != nil {
		walker.currentFileReference.LineReferences[line] = references
		walker.isStartReqSet = true
	}
}

func addLineEnd() {
	// New LineRequirmentEnd
	var references []database.Reference
	for _, ref := range walker.commitReference {
		references = append(references, database.Reference{ref, database.LineRequirmentEnd})
		fmt.Println("Adding LineRequirmentEnd: [reference, line]", ref, walker.currentAddLine-1)
	}
	if references != nil {
		walker.currentFileReference.LineReferences[walker.currentAddLine-1] = references
		walker.isStartReqSet = false
	}
}
