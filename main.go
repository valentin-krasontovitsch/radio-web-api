package main

import (
	"bytes"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	stations map[string]string
	binaries []string
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

func isCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

type Volume struct {
	Value int `json:"volume"`
}

func runCmdAndServe(c *gin.Context, cmd *exec.Cmd) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		errMsg := fmt.Sprint(err) + " - " + stderr.String()
		log.Println(errMsg)
		c.JSON(500, gin.H{"error": errMsg})
	} else {
		log.Println(stdout.String())
		c.JSON(http.StatusOK, gin.H{"output": stdout.String()})
	}
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	// get speaker address
	speaker := os.Getenv("SPEAKER_ADDRESS")
	if speaker == "" {
		speaker = "40:EF:4C:1D:37:F0"
	}
	macRegexp := "^([0-9A-Fa-f]{2}[:-]?){5}([0-9A-Fa-f]{2})$"
	isMacAddress, err := regexp.MatchString(macRegexp, speaker)
	if err != nil {
		log.Fatalf("Failed to check SPEAKER_ADDRESS: %s for mac address", speaker)
	} else if !isMacAddress {
		log.Fatalf("SPEAKER_ADDRESS: %s did not parse as a mac address", speaker)
	}
	log.Println("Speaker: " + speaker)

	binPath := os.Getenv("RWA_BIN_PATH")
	if binPath == "" {
		binPath = "/usr/local/bin"
	}

	// check for binaries
	for _, binary := range binaries {
		if !isCommandAvailable(binary) {
			log.Fatalf("Binary %s could not be found", binary)
		}
	}

	router.GET("/connect", func(c *gin.Context) {
		cmd := exec.Command(binPath + "/connect.sh")
		runCmdAndServe(c, cmd)
	})

	router.GET("/kill", func(c *gin.Context) {
		cmd := exec.Command(binPath + "/kill_radio.sh")
		runCmdAndServe(c, cmd)
	})

	router.GET("/stations", func(c *gin.Context) {
		stats := []string{}
		for name := range stations {
			stats = append(stats, name)
		}
		c.JSON(http.StatusOK, stats)
	})

	router.GET("/play/:station", func(c *gin.Context) {
		// env VOLUME=35\% MUSIC_SOURCE=$BBC2 /usr/local/bin/radio.sh 2>&1
	})

	router.GET("/connected", func(c *gin.Context) {
		cmd := exec.Command(binPath + "/connected.sh")
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			errMsg := fmt.Sprint(err) + " - " + stderr.String()
			log.Println(errMsg)
			c.JSON(http.StatusOK, gin.H{"connected": false})
			return
		}
		log.Println(stdout.String())
		c.JSON(http.StatusOK, gin.H{"connected": true})
	})

	router.GET("/volume", func(c *gin.Context) {
		cmd := exec.Command(binPath + "/get_volume_t5.sh")
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			errMsg := fmt.Sprint(err) + " - " + stderr.String()
			log.Println(errMsg)
			c.JSON(500, gin.H{"error": errMsg})
			return
		}
		volume, convErr := strconv.Atoi(strings.TrimSpace(stdout.String()))
		if convErr != nil {
			log.Println(convErr)
			c.JSON(500, gin.H{"error": convErr})
			return
		}
		log.Println(volume)
		c.JSON(http.StatusOK, gin.H{"volume": volume})
	})

	router.PUT("/volume", func(c *gin.Context) {
		volume := new(Volume)
		jsonErr := c.BindJSON(volume)
		log.Println(volume)
		if jsonErr != nil {
			c.JSON(400, gin.H{"error": jsonErr})
			return
		}
		if volume.Value < 0 || volume.Value > 100 {
			c.JSON(400, gin.H{"error": "Volume should be an integer between" +
				" 0 and 100", "data": volume})
			return
		}
		cmd := exec.Command("set_volume_t5.sh",
			strconv.Itoa(volume.Value)+"%")
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			errMsg := fmt.Sprint(err) + " - " + stderr.String()
			log.Println(errMsg)
			c.JSON(500, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	router.Run() // listen and serve on 0.0.0.0:PORT
}
