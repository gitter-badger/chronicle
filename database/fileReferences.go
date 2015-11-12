package database

import "github.com/libgit2/git2go"

type LineRequirmentType int

const (
	LineRequirmentStart LineRequirmentType = iota
	LineRequirmentEnd
)

type FileReferences struct {
	FileID         git.Oid
	LineReferences map[int][]Reference
}

type Reference struct {
	RequirmentID string
	Type         LineRequirmentType
}
