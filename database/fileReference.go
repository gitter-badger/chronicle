package database

type LineRequirmentType int

const (
	LineRequirmentStart LineRequirmentType = iota
	LineRequirmentEnd
)

type FileReferences struct {
	File           string
	LineReferences map[int][]Reference
	setup          bool
}

func NewFileReferences(file string, lineReferences map[int][]Reference) *FileReferences {
	return &FileReferences{file, lineReferences, true}
}

func (f *FileReferences) Initialized() bool { return f.setup }
func (f *FileReferences) Deinitialize()     { f.setup = false }

type Reference struct {
	RequirmentID string
	Type         LineRequirmentType
}
