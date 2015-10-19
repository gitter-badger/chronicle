package database

import (
	"log"
	"path/filepath"

	"github.com/boltdb/bolt"
)

// Database is a container struct holding a boltDB and gives access to some help function.
type Database struct {
	DB *bolt.DB
}

//NewDatabase returns a database object with given configuration files
func NewDatabase(rootPath string) *Database {
	d := &Database{}
	db, err := bolt.Open("."+string(filepath.Separator)+rootPath+".chronicle"+string(filepath.Separator)+"local.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	d.DB = db
	return d
}

//Close calls the underlying Close()-method on the database object
func (d *Database) Close() error {
	return d.DB.Close()
}
