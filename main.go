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
	BinPath        string   `env:"BIN_PATH" envDefault:"/usr/local/bin"`
	SpeakerAddress string   `env:"SPEAKER_ADDRESS" envDefault:"40:EF:4C:1D:37:F0"`
	Player         string   `env:"PLAYER" envDefault:"/usr/bin/mplayer"`
	PlayerOptions  []string `env:"PLAYER_OPTIONS" envSeparator:"," envDefault:"-really-quiet"`
	AudioControl   string   `env:"AUDIO_CONTROL" envDefault:"/usr/bin/amixer"`
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

func (s session) setVolumeCommand() []string {
	return []string{s.AudioControl, "-q", "sset", "'Master'"}
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
	binaries map[string]string
	version  string
)

func init() {
	binaries = map[string]string{
		"connect":   "connect.sh",
		"connected": "connected.sh",
	}
	stations = make(map[string]string)
	stations["BBC2"] = "http://bbcmedia.ic.llnwd.net/stream/bbcmedia_radio2_mf_p"
	stations["WDR2"] = "https://wdr-wdr2-rheinruhr.sslcast.addradio.de/wdr/wdr2/rheinruhr/mp3/128/stream.mp3"
	stations["NRK-P1"] = "http://lyd.nrk.no/nrk_radio_p1_hordaland_mp3_h"
	stations["NRK-P3"] = "http://lyd.nrk.no/nrk_radio_p3_mp3_h"
	stations["RADIO-NORGE"] = "http://live-bauerno.sharp-stream.com/radionorge_no_mp3"
	stations["xmas"] = "http://live-bauerno.sharp-stream.com/station17_no_hq"
}

func runCmdAndServe(c *gin.Context, cmd []string, env map[string]string) {
	stdout, _, err := runCommand(cmd, env)
	if err != nil {
		errMsg := fmt.Sprintf("%+v", err)
		log.Println(errMsg)
		c.JSON(500, gin.H{"error": errMsg})
		return
	}
	if stdout == "" {
		c.Status(http.StatusNoContent)
		return
	}
	log.Println(stdout)
	c.JSON(http.StatusOK, gin.H{"output": stdout})
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
	cleanStdout := strings.TrimSpace(stdoutB.String())
	stderrString := stderrB.String()
	if err != nil {
		err = errors.Wrap(err, stderrString)
		err = errors.WithMessage(err, cleanStdout)
		err = errors.WithMessage(err, "Cmd run: "+strings.Join(cmdArgs, " "))
	}
	return cleanStdout, stderrB.String(), err
}

func handleError(c *gin.Context, status int, err error) {
	errMsg := fmt.Sprintf("%+v", err)
	log.Println(errMsg)
	c.JSON(status, gin.H{"error": errMsg})
}

func parseVolume(stdout string) (int, error) {
	log.Println(stdout)
	lines := strings.Split(stdout, "\n")
	lastLine := lines[len(lines)-1]
	log.Println(lastLine)
	volumeRegexp := regexp.MustCompile(`\[[0-9]+%\]`)
	bracketedVolume := volumeRegexp.FindString(lastLine)
	volumeString := strings.Trim(bracketedVolume, "[]%")
	volume, convErr := strconv.Atoi(volumeString)
	return volume, convErr
}

func (s session) getVolume() (v Volume, err error) {
	cmd := []string{s.AudioControl, "sget", "'Master'"}
	stdout, _, err := runCommand(cmd, nil)
	if err != nil {
		return
	}
	volume, err := parseVolume(stdout)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	v = Volume{Value: volume}
	return
}

func (s session) connect(c *gin.Context) {
	cmd := []string{filepath.Join(s.BinPath, binaries["connect"]),
		s.SpeakerAddress}
	runCmdAndServe(c, cmd, nil)
}

func (s session) disconnect(c *gin.Context) {
	cmd := []string{"bluetoothctl", "disconnect", s.SpeakerAddress}
	runCmdAndServe(c, cmd, nil)
}

func (s session) killerHandler(c *gin.Context) {
	killIt := []string{"killall", "-q", s.Player}
	runCommand(killIt, nil)
	c.Status(http.StatusNoContent)
}

func getStations(c *gin.Context) {
	stats := []string{}
	for name := range stations {
		stats = append(stats, name)
	}
	c.JSON(http.StatusOK, stats)
}

func (s session) playStation(c *gin.Context) {
	station := c.Param("station")
	url, ok := stations[station]
	if !ok {
		errMsg := fmt.Sprintf("Do not have a station %s", station)
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}
	currentVolume, err := s.getVolume()
	if err != nil {
		errMsg := fmt.Sprintf("%+v", err)
		log.Println(errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	if currentVolume.Value > 60 {
		difference := currentVolume.Value - 45
		err := s.ChangeVolume(difference, false)
		if err != nil {
			errMsg := fmt.Sprintf("%+v", err)
			log.Println(errMsg)
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
			return
		}
	}
	playArgs := append(s.PlayerOptions, url)
	proc := exec.Command(s.Player, playArgs...)
	err = proc.Start()
	if err != nil {
		err = errors.WithStack(err)
		errMsg := fmt.Sprintf("%+v", err)
		log.Println(errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	err = proc.Process.Release()
	if err != nil {
		err = errors.WithStack(err)
		errMsg := fmt.Sprintf("%+v", err)
		log.Println(errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	c.Status(http.StatusNoContent)
}

func evaluateStdout(stdout, truthy, falsy string) (bool, error) {
	if stdout == truthy {
		return true, nil
	} else if stdout == falsy {
		return false, nil
	} else {
		err := errors.Errorf("Expected `%s` or `%s`, got `%s` instead", truthy,
			falsy, stdout)
		return false, err
	}
}

func (s session) connected() (connected bool, err error) {
	cmd := []string{filepath.Join(s.BinPath, binaries["connected"]),
		s.SpeakerAddress}
	_, stderr, err := runCommand(cmd, nil)
	if err != nil && stderr != "" {
		return
	}
	return err == nil, nil
}

func (s session) connectedHandler(c *gin.Context) {
	connected, err := s.connected()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"connected": connected})
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
	cmd := append(s.setVolumeCommand(),
		fmt.Sprintf("%d%%%s", amount, sign))
	_, _, err := runCommand(cmd, nil)
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
	cmd := append(s.setVolumeCommand(), "toggle")
	_, _, err := runCommand(cmd, nil)
	if err != nil {
		errMsg := fmt.Sprintf("%+v", err)
		log.Println(errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s session) muted() (bool, error) {
	cmd := []string{s.AudioControl, "sget", "'Master'"}
	stdout, _, err := runCommand(cmd, nil)
	if err != nil {
		return false, err
	}
	lines := strings.Split(stdout, "\n")
	lastLine := lines[len(lines)-1]
	re := regexp.MustCompile(`\[(on|off)\]`)
	stateBrackets := re.FindString(lastLine)
	state := strings.Trim(stateBrackets, "[]")
	return evaluateStdout(state, "off", "on")
}

func (s session) mutedHandler(c *gin.Context) {
	muted, err := s.muted()
	if err != nil {
		errMsg := fmt.Sprintf("%+v", err)
		log.Println(errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"muted": muted})
}

func (s session) statusHandler(c *gin.Context) {
	connected, err := s.connected()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	if !connected {
		c.JSON(http.StatusOK, gin.H{"connected": connected})
		return
	}
	volume, err := s.getVolume()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	muted, err := s.muted()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"connected": connected, "volume": volume,
		"muted": muted})
}

func explainAPI(c *gin.Context) {
	docstring := `radio web api endpoints
  /connect         - connect to the bluetooth speaker
  /disconnect      - disconnect the bluetooth speaker
  /kill            - kills players
  /stations        - list available stations. returns ` + "`[[STRING], [STRING], ...]`" + `
  /play/:station   - starts playing station, expects one of names returned by
                       /stations endpoint
  /connected       - checks whether we are connected. returns ` + "`{\"connected\": [BOOL]}`" + `
  /volume          - get volume. returns ` + "`{\"volume\": [INT]}`" + `
  /mute            - mutes or unmutes the radio
  /muted           - whether we are muted, or not. returns ` + "`{\"muted\": [BOOL]}`" + `
  /louder/:amount  - increases volume by non-negative int amount
  /quiet/:amount   - decreases volume by non-negative int amount
  /version         - version of this deployment. returns ` + "`{\"volume\": [STRING]}`" + `
  /status          - returns  ` + "`{\"connected\": false}`" + ` if not connected, and
                     ` + "`{\"connected\": true, \"volume\": [STRING], \"muted\": [BOOL]}`" + `
  /                - returns this documentation
`
	c.String(http.StatusOK, docstring)
}

func setupRouter(s session) *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/connect", s.connect)
	router.GET("/disconnect", s.disconnect)
	router.GET("/kill", s.killerHandler)
	router.GET("/stations", getStations)
	router.GET("/play/:station", s.playStation)
	router.GET("/connected", s.connectedHandler)
	router.GET("/volume", s.getVolumeHandler)
	router.GET("/mute", s.mute)
	router.GET("/muted", s.mutedHandler)
	router.GET("/louder/:amount", createVolumeChangerHandler(s, true))
	router.GET("/quiet/:amount", createVolumeChangerHandler(s, false))
	router.GET("/status", s.statusHandler)
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
