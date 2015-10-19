package requirments

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Benefactory/chronicle/database"
	"github.com/naoina/toml"
)

// ParseReqFile parse a specific .req file formated with .toml format
func ParseReqFile(path string, db *database.Database) error {
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

	// fmt.Println("Relationships", requirments.Relationship)

	for key, value := range requirments.Feature {
		fmt.Println(key, value.Status)
		// fmt.Println("Relationships", value.Relationship)
		// for _, relationship := range value.Relationship {
		// 	fmt.Println(relationship.Status)
		// }

	}
	return nil
}
