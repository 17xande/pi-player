package main

import (
	"fmt"
	"os"
	"os/exec"
)

func testPlay() {
	cmd := exec.Command("omxplayer", "/home/pi/movies/Bee Movie.mp4", "-b")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	fmt.Println("Process ID: ", cmd.Process.Pid)
}
