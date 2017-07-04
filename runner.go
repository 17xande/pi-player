package main

import (
	"os"
	"os/exec"
)

func play() {
	cmd := exec.Command("omxplayer", "~/movies/Bee Movie.mp4", "-b")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
