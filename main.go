package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type Volume struct {
	Value int `json:"volume"`
}

type session struct {
	BinPath        string `env:"BIN_PATH" envDefault:"/usr/local/bin"`
	SpeakerAddress string `env:"SPEAKER_ADDRESS" envDefault:"40:EF:4C:1D:37:F0"`
}

func initSession() (session, error) {
	s := session{}
	err := env.Parse(&s)
	if err != nil {
		return s, errors.WithStack(err)
	}
	macRegexp := "^([0-9A-Fa-f]{2}[:-]?){5}([0-9A-Fa-f]{2})$"
	isMacAddress, err := regexp.MatchString(macRegexp, s.SpeakerAddress)
	if err != nil {
		return s, errors.WithStack(err)
	} else if !isMacAddress {
		return s, errors.WithStack(err)
	}

	for _, binary := range binaries {
		if !s.isCommandAvailable(binary) {
			log.Fatalf("Binary %s could not be found", binary)
		}
	}
	return s, nil
}

func (s session) isCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

var (
	stations map[string]string
	binaries []string
	version  string
)

func init() {
	binaries = []string{
		"connect.sh",
		"connected.sh",
		"set_volume_t5.sh",
		"get_volume_t5.sh",
		"kill_radio.sh",
		"radio.sh",
	}
	stations = make(map[string]string)
	stations["BBC2"] = "http://bbcmedia.ic.llnwd.net/stream/bbcmedia_radio2_mf_p"
	stations["WDR2"] = "https://wdr-wdr2-rheinruhr.sslcast.addradio.de/wdr/wdr2/rheinruhr/mp3/128/stream.mp3"
	stations["NRK-P1"] = "http://lyd.nrk.no/nrk_radio_p1_hordaland_mp3_h"
	stations["NRK-P3"] = "http://lyd.nrk.no/nrk_radio_p3_mp3_h"
	stations["RADIO-NORGE"] = "http://live-bauerno.sharp-stream.com/radionorge_no_mp3"
	stations["xmas"] = "http://live-bauerno.sharp-stream.com/station17_no_hq"
}

func runCmdAndServe(c *gin.Context, cmd string, env map[string]string) {
	stdout, stderr, err := runCommand([]string{cmd}, env)
	if err != nil {
		err = errors.Wrap(err, stderr)
		errMsg := fmt.Sprintf("%+v", err)
		log.Println(errMsg)
		c.JSON(500, gin.H{"error": errMsg})
	} else {
		log.Println(stdout)
		c.JSON(http.StatusOK, gin.H{"output": stdout})
	}
}

func makeEnvAssignments(env map[string]string) []string {
	assignments := []string{}
	for key, value := range env {
		assignments = append(assignments, fmt.Sprintf("%s=%s", key, value))
	}
	return assignments
}

func runCommand(cmdArgs []string, env map[string]string) (stdout string, stderr string, err error) {
	binary := cmdArgs[0]
	cmd := exec.Command(binary, cmdArgs[1:]...)
	var stdoutB bytes.Buffer
	var stderrB bytes.Buffer
	cmd.Stdout = &stdoutB
	cmd.Stderr = &stderrB
	envAssignments := makeEnvAssignments(env)
	cmd.Env = append(os.Environ(), envAssignments...)
	err = cmd.Run()
	return stdoutB.String(), stderrB.String(), err
}

func (s session) getVolume() (v Volume, err error) {
	cmd := filepath.Join(s.BinPath, "get_volume_t5.sh")
	stdout, stderr, err := runCommand([]string{cmd}, nil)
	if err != nil {
		err = errors.Wrap(err, stderr)
		log.Printf("%+v", err)
		return
	}
	volumeNumber, convErr := strconv.Atoi(strings.TrimSpace(stdout))
	if convErr != nil {
		err = errors.WithStack(convErr)
		log.Printf("%+v", err)
		return
	}
	v = Volume{Value: volumeNumber}
	return
}

func (s session) connect(c *gin.Context) {
	cmd := filepath.Join(s.BinPath, "connect.sh")
	runCmdAndServe(c, cmd, nil)
}

func (s session) killerHandler(c *gin.Context) {
	cmd := filepath.Join(s.BinPath, "kill_radio.sh")
	runCmdAndServe(c, cmd, nil)
}

func getStations(c *gin.Context) {
	stats := []string{}
	for name := range stations {
		stats = append(stats, name)
	}
	c.JSON(http.StatusOK, stats)
}

func (s session) playStation(c *gin.Context) {
	// TODO set max trials to 1 or improve kill script!
	station := c.Param("station")
	url, ok := stations[station]
	if !ok {
		errMsg := fmt.Sprintf("Do not have a station %s", station)
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}
	cmd := filepath.Join(s.BinPath, "radio.sh")
	env := map[string]string{}
	env["MUSIC_SOURCE"] = url
	currentVolume, err := s.getVolume()
	if err != nil {
		errMsg := fmt.Sprintf("%+v", err)
		log.Println(errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	} else {
		if currentVolume.Value > 60 {
			env["VOLUME"] = "45%"
		}
	}
	_, stderr, err := runCommand([]string{cmd}, env)
	if err != nil {
		err = errors.Wrap(err, stderr)
		errMsg := fmt.Sprintf("%+v", err)
		log.Println(errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s session) connectedHandler(c *gin.Context) {
	cmd := filepath.Join(s.BinPath, "connected.sh")
	_, stderr, err := runCommand([]string{cmd}, nil)
	if err != nil {
		err = errors.Wrap(err, stderr)
		errMsg := fmt.Sprintf("%+v", err)
		log.Println(errMsg)
		c.JSON(http.StatusOK, gin.H{"connected": false, "error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"connected": true})
}

func (s session) getVolumeHandler(c *gin.Context) {
	volume, err := s.getVolume()
	if err != nil {
		errMsg := fmt.Sprintf("%+v", err)
		log.Println(errMsg)
		c.JSON(500, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, volume)
}

func (s session) ChangeVolume(amount int, louder bool) error {
	sign := "-"
	if louder {
		sign = "+"
	}
	cmd := []string{filepath.Join(s.BinPath, "set_volume_t5.sh"),
		fmt.Sprintf("%d%%%s", amount, sign)}
	_, stderr, err := runCommand(cmd, nil)
	if err != nil {
		err = errors.Wrap(err, stderr)
	}
	return err
}

func createVolumeChangerHandler(s session, louder bool) func(c *gin.Context) {
	handler := func(c *gin.Context) {
		amountString := c.Param("amount")
		amount, err := strconv.Atoi(amountString)
		if err != nil {
			errMsg := fmt.Sprintf("%+v", errors.WithStack(err))
			log.Println(errMsg)
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		err = s.ChangeVolume(amount, louder)
		if err != nil {
			errMsg := fmt.Sprintf("%+v", err)
			log.Println(errMsg)
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
			return
		}
		c.Status(http.StatusNoContent)
	}
	return handler
}

func (s session) mute(c *gin.Context) {
	cmd := []string{filepath.Join(s.BinPath, "set_volume_t5.sh"), "toggle"}
	_, stderr, err := runCommand(cmd, nil)
	if err != nil {
		errMsg := fmt.Sprintf("%+v", errors.Wrap(err, stderr))
		log.Println(errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	c.Status(http.StatusNoContent)
}

func explainAPI(c *gin.Context) {
	docstring := `radio web api endpoints
  /connect         - connect to the bluetooth speaker
  /kill            - kills players
  /stations        - returns list of available stations
  /play/:station   - starts playing station, expects one of names returned by
                       /stations endpoint
  /connected       - checks whether we are connected
  /volume          - get volume
  /mute            - mutes or unmutes the radio
  /louder/:amount  - increases volume by non-negative int amount
  /quiet/:amount   - decreases volume by non-negative int amount
  /version         - version of this deployment
  /                - returns this documentation
`
	c.String(http.StatusOK, docstring)
}

func setupRouter(s session) *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/connect", s.connect)
	router.GET("/kill", s.killerHandler)
	router.GET("/stations", getStations)
	router.GET("/play/:station", s.playStation)
	router.GET("/connected", s.connectedHandler)
	router.GET("/volume", s.getVolumeHandler)
	router.GET("/mute", s.mute)
	router.GET("/louder/:amount", createVolumeChangerHandler(s, true))
	router.GET("/quiet/:amount", createVolumeChangerHandler(s, false))
	router.GET("/", explainAPI)
	router.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, version+"\n")
	})
	return router
}

func main() {
	s, err := initSession()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	router := setupRouter(s)
	router.Run() // listen and serve on 0.0.0.0:PORT
}
