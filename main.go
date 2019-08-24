package main

import (
	"flag"
	"io/ioutil"
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
	ticktime    *int

	ticker *time.Ticker

	firstTick = true

	lastPingOk     bool
	lastCifsOk     bool
	lastReadfileOk bool
)

func init() {
	common.Init("cifsmon", "1.0.0", "2018", "monitor the accessibility of shares", "mpetavy", common.APACHE, "https://github.com/mpetavy/traclink", true, nil, nil, tick, time.Second*5)

	host = flag.String("h", "", "host")
	port = flag.Int("p", 445, "cifs port")
	user = flag.String("u", "", "cifs user")
	domain = flag.String("d", "", "cifs domain")
	workstation = flag.String("s", "", "cifs workstation")
	password = flag.String("w", "", "cifs password")
	filename = flag.String("f", "", "filename")
	ticktime = flag.Int("t", 1000, "ticktime in ms")
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
	debug := false
	session, err := smb.NewSession(options, debug)
	if err != nil {
		common.DebugError(err)

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

	return err == nil
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
		common.DebugError(err)
		common.Debug("Pinging failed")
	}

	return err == nil
}

func readfile() bool {
	if *filename == "" {
		return true
	}

	common.Debug("Trying to read file %s ...", *filename)

	b, err := common.FileExists(*filename)
	if b {
		_, err = ioutil.ReadFile(*filename)

		b = err == nil
	}

	if b {
		common.Debug("Reading file %s successful", *filename)
	} else {
		common.DebugError(err)
		common.Debug("Reading file %s failed", *filename)
	}

	return b
}

func tick() error {
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
