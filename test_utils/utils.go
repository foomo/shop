package test_utils

import (
	"log"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/foomo/shop/configuration"
)

func GetTestUtilsDir() string {
	_, filename, _, _ := runtime.Caller(1)
	filename = strings.Replace(filename, "/test_utils.go", "", -1) // remove "utils.go"
	return path.Dir(filename)                                      // remove //"utils" and return
}

// Drops order collection and event_log collection
func DropAllCollections() error {
	db := ""
	if strings.Contains(configuration.MONGO_URL, configuration.LocalUnitTests) {
		db = configuration.LocalUnitTests
	} else {
		db = configuration.WithDocker
	}
	cmd := exec.Command("mongo", db, GetTestUtilsDir()+"/dropCollections.js")
	log.Println("Command.args: ", cmd.Args)
	//cmd := exec.Command("mongo", "localhost:27017/"+foomo_shop_config.MONGO_DB, GetProjectDir()+"/mongo/dropCollections.js")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Printf("Waiting for command to finish...")
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Printf("Command finished with error: %v", err)
	return nil
}
