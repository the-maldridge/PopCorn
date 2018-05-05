package main

import (
	"os/exec"
	"log"
	"bytes"
	"flag"
	"time"
	"os"
	"io/ioutil"
	"math/rand"
)

var (
	uuidPath = flag.String("uuid_path", "/etc/popcorn/uuid", "Path to the uuid file")
	machineID = []byte{}
)

func getUUID() []byte {
	ID, err := ioutil.ReadFile(*uuidPath)
	if os.IsNotExist(err) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		ID = make([]byte, 96)
		log.Println(r.Read(ID))
		log.Println(ID)
		if err := ioutil.WriteFile(*uuidPath, ID, 0644); err != nil {
			log.Fatal(err)
		}
	}
	return ID
}

func main() {
	flag.Parse()
	machineID = getUUID()
	log.Printf("Machine ID: %s", machineID)

	_, err := exec.LookPath("xbps-query")
	if err != nil {
		log.Println("xbps-query isn't in $PATH.  Are you sure this is a Void system?")
		log.Fatal(err)
	}

	xbpsQueryCmd := exec.Command("xbps-query", "-m")
	var out bytes.Buffer
	xbpsQueryCmd.Stdout = &out
	if err := xbpsQueryCmd.Run(); err != nil {
		log.Fatal(err)
	}
	log.Println(out.String())

	_, err = exec.LookPath("xuname")
	if err != nil {
		return
	}
	xunameCmd := exec.Command("xuname")
	xunameCmd.Stdout = &out
	if err := xunameCmd.Run(); err != nil {
		log.Fatal(err)
	}
	log.Println(out.String())
}
