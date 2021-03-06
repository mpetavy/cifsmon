package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/mpetavy/common"
	"github.com/stacktitan/smb/smb"
)

var (
	host        *string
	port        *int
	user        *string
	domain      *string
	workstation *string
	password    *string
	filename    *string

	firstTick = true

	lastPingOk     bool
	lastCifsOk     bool
	lastReadfileOk bool
)

func init() {
	common.Init(true, "1.0.0", "", "", "2018", "monitor the accessibility of shares", "mpetavy", fmt.Sprintf("https://github.com/mpetavy/%s", common.Title()), common.APACHE, nil, nil, nil, run, time.Second*5)

	host = flag.String("h", "", "host")
	port = flag.Int("p", 445, "cifs port")
	user = flag.String("u", "", "cifs user")
	domain = flag.String("d", "", "cifs domain")
	workstation = flag.String("s", "", "cifs workstation")
	password = flag.String("w", "", "cifs password")
	filename = flag.String("f", "", "filename")
}

func cifs() bool {
	common.Debug("Trying CIFS login to %s ...", *host)

	options := smb.Options{
		Host:        *host,
		Port:        *port,
		User:        *user,
		Domain:      *domain,
		Workstation: *workstation,
		Password:    *password,
	}
	session, err := smb.NewSession(options, false)
	if err != nil {
		common.Error(err)

		return false
	}
	defer session.Close()

	if session.IsSigningRequired {
		common.Debug("Signing is required")
	} else {
		common.Debug("Signing is not required")
	}

	if session.IsAuthenticated {
		common.Debug("Login successful")
	} else {
		common.Debug("Login failed")
	}

	return true
}

func ping() bool {
	var err error

	common.Debug("Trying to ping %s ...", *host)

	if common.IsWindowsOS() {
		cmd := exec.Command("ping", "-n", "1", *host)

		err = cmd.Run()
	} else {
		cmd := exec.Command("ping", "-c", "1", *host)

		err = cmd.Run()
	}

	if err == nil {
		common.Debug("Pinging successful")
	} else {
		common.Error(err)
		common.Debug("Pinging failed")
	}

	return err == nil
}

func readfile() bool {
	if *filename == "" {
		return true
	}

	common.Debug("Trying to read file %s ...", *filename)

	b := common.FileExists(*filename)
	if b {
		_, err := os.ReadFile(*filename)

		b = err == nil

		if !b {
			common.Error(err)
		}
	}

	if b {
		common.Debug("Reading file %s successful", *filename)
	} else {
		common.Debug("Reading file %s failed", *filename)
	}

	return b
}

func run() error {
	pingOk := ping()
	cifsOk := cifs()
	readfileOk := readfile()

	if firstTick || (pingOk != lastPingOk) || (cifsOk != lastCifsOk) || (readfileOk != lastReadfileOk) {
		if firstTick {
			common.Info("-- Initial status --")
		} else {
			common.Info("-- Status changed --")
		}

		if firstTick || pingOk != lastPingOk {
			common.Info("Ping: %v", pingOk)
		}

		if firstTick || cifsOk != lastCifsOk {
			common.Info("Cifs: %v", cifsOk)
		}

		if *filename != "" && (firstTick || readfileOk != lastReadfileOk) {
			common.Info("Read file: %v", readfileOk)
		}

		lastPingOk = pingOk
		lastCifsOk = cifsOk
		lastReadfileOk = readfileOk
	}

	firstTick = false

	return nil
}

func main() {
	defer common.Done()

	common.Run([]string{"h", "u", "w", "d"})
}
