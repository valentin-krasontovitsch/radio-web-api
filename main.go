package main

import (
	"bytes"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

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
		c.JSON(200, gin.H{"output": stdout.String()})
	}
}

func main() {
	router := gin.Default()
	speaker := os.Getenv("SPEAKER_ADDRESS")
	log.Println("Speaker: " + speaker)

	// TODO read from environment
	binPath := "/home/val/code/blue-radio-shell"
	log.Println(binPath)

	router.Use(cors.Default())

	router.GET("/connect", func(c *gin.Context) {
		cmd := exec.Command(binPath + "/connect.sh")
		runCmdAndServe(c, cmd)
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
			c.JSON(200, gin.H{"connected": false})
			return
		}
		log.Println(stdout.String())
		c.JSON(200, gin.H{"connected": true})
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
		c.JSON(200, gin.H{"volume": volume})
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
		cmd := exec.Command(binPath+"/set_volume_t5.sh",
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
		c.JSON(200, gin.H{})
	})

	router.Run() // listen and serve on 0.0.0.0:PORT
}
