package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/the-maldridge/popcorn/pkg/stats"
)

var (
	appLogger hclog.Logger
	machineID string

	backoff = []time.Duration{
		5 * time.Second,
		5 * time.Second,
		5 * time.Second,
		10 * time.Second,
		10 * time.Second,
		10 * time.Second,
		1 * time.Minute,
		1 * time.Minute,
		1 * time.Minute,
	}
)

func getUUID(path string) (string, error) {
	ID, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		ID = make([]byte, 96)
		r.Read(ID)
		if err := ioutil.WriteFile(path, ID, 0644); err != nil {
			return "", err
		}
	}
	h := sha256.Sum256(ID)
	return fmt.Sprintf("%x", h), nil
}

func getPkgs() ([]stats.Package, error) {
	_, err := exec.LookPath("xbps-query")
	if err != nil {
		return nil, errors.New("xbps-query isn't in $PATH.  Are you sure this is a Void system?")
	}

	xbpsQueryCmd := exec.Command("xbps-query", "-m")
	var out bytes.Buffer
	xbpsQueryCmd.Stdout = &out
	if err := xbpsQueryCmd.Run(); err != nil {
		return nil, err
	}

	pkgs := []stats.Package{}
	for _, p := range strings.Split(out.String(), "\n") {
		parts := strings.Split(p, "-")
		pkg := stats.Package{
			Name:    strings.Join(parts[:len(parts)-1], "-"),
			Version: parts[len(parts)-1],
		}
		if pkg.Name == "" {
			continue
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

func getXUname() (stats.XUname, error) {
	var out bytes.Buffer
	_, err := exec.LookPath("xuname")
	if err != nil {
		return stats.XUname{}, err
	}
	xunameCmd := exec.Command("xuname")
	xunameCmd.Stdout = &out
	if err := xunameCmd.Run(); err != nil {
		log.Fatal(err)
	}
	fields := strings.Fields(out.String())

	return stats.XUname{
		OSName:       fields[0],
		Kernel:       fields[1],
		Mach:         fields[2],
		CPUInfo:      fields[3],
		UpdateStatus: fields[4],
		RepoStatus:   fields[5],
	}, nil
}

func collectAndSendStats() {
	pkgs, err := getPkgs()
	if err != nil {
		appLogger.Error("Could not obtain package list", "error", err)
		return
	}
	xuname, err := getXUname()
	if err != nil {
		appLogger.Warn("Could not obtain xuname output", "error", err)
	}
	d := stats.Stats{
		Packages: pkgs,
		XUname:   xuname,
	}

	body, err := json.Marshal(d)
	if err != nil {
		appLogger.Error("Could not marshal data", "error", err)
		return
	}

	c := &http.Client{
		Timeout: time.Second * 15,
	}

	req, err := http.NewRequest(http.MethodPost, os.Getenv("STATS_URL"), bytes.NewBuffer(body))
	if err != nil {
		appLogger.Error("Could not prepare request", "error", err)
		return
	}
	req.Header.Add("From", machineID)
	req.Header.Add("Content-type", "application/json")

	for _, td := range backoff {
		resp, err := c.Do(req)
		if err != nil {
			appLogger.Warn("Error while transmitting stats data", "error", err)
			time.Sleep(td)
			continue
		}
		srvMsg, _ := ioutil.ReadAll(resp.Body)
		appLogger.Info("Message from server", "message", string(srvMsg))
		appLogger.Info("Stats updated")
		return
	}
	appLogger.Error("Could not transmit stats, giving up")
}

func main() {
	llevel := os.Getenv("LOG_LEVEL")
	if llevel == "" {
		llevel = "INFO"
	}
	appLogger = hclog.New(&hclog.LoggerOptions{
		Name:  "popcorn",
		Level: hclog.LevelFromString(llevel),
	})

	idPath := os.Getenv("UUID_PATH")
	if idPath == "" {
		idPath = "/etc/popcorn/uuid"
	}
	var err error
	machineID, err = getUUID(idPath)
	if err != nil {
		appLogger.Error("Could not obtain unique identity", "error", err)
		os.Exit(1)
	}

	collectAndSendStats()
}
