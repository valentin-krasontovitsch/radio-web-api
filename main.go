package main

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/exec"
)

func runCmdAndServe(c *gin.Context, cmd *exec.Cmd) {
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		c.JSON(503, gin.H{"error": err})
	} else {
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

	r.GET("/connect", func(c *gin.Context) {
		cmd := exec.Command(binPath + "/connect.sh")
		cmd.Env = append(os.Environ(), "CONNECT_TRIALS=2")
		runCmdAndServe(c, cmd)
	})

	r.GET("/connected", func(c *gin.Context) {
		cmd := exec.Command(binPath + "/connected.sh")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		log.Println(out.String())
		log.Println(err)
		if err != nil {
			c.JSON(200, gin.H{"connected": "false"})
		} else {
			c.JSON(200, gin.H{"connected": "true"})
		}
	})
	r.Run() // listen and serve on 0.0.0.0:PORT
}
