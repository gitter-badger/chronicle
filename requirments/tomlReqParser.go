package requirments

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/Benefactory/chronicle/database"
	"github.com/boltdb/bolt"
	"github.com/naoina/toml"
)

// ParseReqFile parse a specific .req file formated with .toml format
func ParseReqFile(path string, db *database.Database, date time.Time) error {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	var requirments tomlReq
	if err := toml.Unmarshal(buf, &requirments); err != nil {
		panic(err)
	}

	for _, value := range requirments.Component {
		err = db.DB.Update(func(tx *bolt.Tx) error {
			bRoot := tx.Bucket([]byte("RootBucket"))
			bCurrent := bRoot.Bucket([]byte(date.Format(time.RFC3339)))
			err := bCurrent.Put([]byte("title_"+value.Id), []byte(value.Title))
			err = bCurrent.Put([]byte("type_"+value.Id), []byte("Component"))
			// TODO: JSON insert of the whole object
			return err
		})
		if err != nil {
			panic(err)
		}
	}

	for _, value := range requirments.Feature {
		err = db.DB.Update(func(tx *bolt.Tx) error {
			bRoot := tx.Bucket([]byte("RootBucket"))
			b := bRoot.Bucket([]byte(date.Format(time.RFC3339)))
			err := b.Put([]byte("title_"+value.Id), []byte(value.Title))
			err = b.Put([]byte("type_"+value.Id), []byte("Feature"))
			// TODO: JSON insert of the whole object
			return err
		})
		if err != nil {
			panic(err)
		}
	}

	for _, value := range requirments.Format {
		err = db.DB.Update(func(tx *bolt.Tx) error {
			bRoot := tx.Bucket([]byte("RootBucket"))
			b := bRoot.Bucket([]byte(date.Format(time.RFC3339)))
			err := b.Put([]byte("title_"+value.Id), []byte(value.Title))
			err = b.Put([]byte("type_"+value.Id), []byte("Format"))
			// TODO: JSON insert of the whole object
			return err
		})
		if err != nil {
			panic(err)
		}
	}

	for _, value := range requirments.Function {
		err = db.DB.Update(func(tx *bolt.Tx) error {
			bRoot := tx.Bucket([]byte("RootBucket"))
			b := bRoot.Bucket([]byte(date.Format(time.RFC3339)))
			err := b.Put([]byte("title_"+value.Id), []byte(value.Title))
			err = b.Put([]byte("type_"+value.Id), []byte("Function"))
			// TODO: JSON insert of the whole object
			return err
		})
		if err != nil {
			panic(err)
		}
	}

	for _, value := range requirments.Goal {
		err = db.DB.Update(func(tx *bolt.Tx) error {
			bRoot := tx.Bucket([]byte("RootBucket"))
			b := bRoot.Bucket([]byte(date.Format(time.RFC3339)))
			err := b.Put([]byte("title_"+value.Id), []byte(value.Title))
			err = b.Put([]byte("type_"+value.Id), []byte("Goal"))
			// TODO: JSON insert of the whole object
			return err
		})
		if err != nil {
			panic(err)
		}
	}

	for _, value := range requirments.Module {
		err = db.DB.Update(func(tx *bolt.Tx) error {
			bRoot := tx.Bucket([]byte("RootBucket"))
			b := bRoot.Bucket([]byte(date.Format(time.RFC3339)))
			err := b.Put([]byte("title_"+value.Id), []byte(value.Title))
			err = b.Put([]byte("type_"+value.Id), []byte("Module"))
			// TODO: JSON insert of the whole object
			return err
		})
		if err != nil {
			panic(err)
		}
	}

	for _, value := range requirments.Risk {
		err = db.DB.Update(func(tx *bolt.Tx) error {
			bRoot := tx.Bucket([]byte("RootBucket"))
			b := bRoot.Bucket([]byte(date.Format(time.RFC3339)))
			err := b.Put([]byte("title_"+value.Id), []byte(value.Title))
			err = b.Put([]byte("type_"+value.Id), []byte("Risk"))
			// TODO: JSON insert of the whole object
			return err
		})
		if err != nil {
			panic(err)
		}
	}

	for _, value := range requirments.Stakeholder {
		err = db.DB.Update(func(tx *bolt.Tx) error {
			bRoot := tx.Bucket([]byte("RootBucket"))
			b := bRoot.Bucket([]byte(date.Format(time.RFC3339)))
			err := b.Put([]byte("title_"+value.Id), []byte(value.Title))
			err = b.Put([]byte("type_"+value.Id), []byte("Stakeholder"))
			// TODO: JSON insert of the whole object
			return err
		})
		if err != nil {
			panic(err)
		}
	}
	return nil
}
