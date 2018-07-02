package main

import (
	"bytes"
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
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": err})
	} else {
		log.Println(out.String())
		c.JSON(200, gin.H{"output": out.String()})
	}
}

func main() {
	r := gin.Default()
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

	r.GET("/connected", func(c *gin.Context) {
		cmd := exec.Command(binPath + "/connected.sh")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			log.Println(err)
			c.JSON(200, gin.H{"connected": false})
			return
		}
		log.Println(out.String())
		c.JSON(200, gin.H{"connected": true})
	})

	r.GET("/volume", func(c *gin.Context) {
		cmd := exec.Command(binPath + "/get_volume_t5.sh")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{"error": "Error running command: " + err.Error()})
			return
		}
		volume, convErr := strconv.Atoi(strings.TrimSpace(out.String()))
		if convErr != nil {
			log.Println(convErr)
			c.JSON(500, gin.H{"error": convErr})
			return
		}
		log.Println(volume)
		c.JSON(200, gin.H{"volume": volume})
	})

	r.PUT("/volume", func(c *gin.Context) {
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
		err := cmd.Run()
		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{"error": err})
			return
		}
		c.JSON(200, gin.H{})
	})

	r.Run() // listen and serve on 0.0.0.0:PORT
}
