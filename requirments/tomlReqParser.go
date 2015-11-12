package requirments

import (
	"time"

	"github.com/Benefactory/chronicle/database"
	"github.com/boltdb/bolt"
	"github.com/naoina/toml"
)

// ParseReqFile parse a specific .req file formated with .toml format
func ParseReqFile(data []byte, db *database.Database, date time.Time) error {
	var requirments tomlReq

	err := toml.Unmarshal(data, &requirments)
	if err != nil {
		return err
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
