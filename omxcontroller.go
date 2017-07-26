package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// Player represents the omxplayer
type Player struct {
	command *exec.Cmd
	pipeIn  io.WriteCloser
	debug   bool
}

var commandList = map[string]string{
	"speedIncrease":   "1",
	"speedDecrease":   "2",
	"rewind":          "<",
	"fastForward":     ">",
	"chatperPrevious": "i",
	"chapterNext":     "o",
	"exit":            "q",
	"quit":            "q",
	"pauseResume":     "p",
	"volumeDecrease":  "-",
	"volumeIncrease":  "+",
	"seekBack30":      "\x1b[D",
	"seekForward30":   "\x1b[C",
	"seekBack600":     "\x1b[B",
	"seekForward600":  "\x1b[A",
}

// Start starts the player
func (p *Player) Start(path string, position time.Duration) error {
	var err error
	pos := fmt.Sprintf("%d:%d:%d", int(position.Hours()), int(position.Minutes()), int(position.Seconds()))

	if p.debug {
		fmt.Println("omxplayer -b -l", pos, path)
		p.pipeIn = os.Stdout
		return nil
	}
	p.command = exec.Command("omxplayer", "-b", "-l", pos, path)
	p.pipeIn, err = p.command.StdinPipe()

	if err != nil {
		return err
	}

	p.command.Stdout = os.Stdout
	err = p.command.Start()
	return err
}

// SendCommand sends a command to the omxplayer process
func (p *Player) SendCommand(command string) error {
	cmd, ok := commandList[command]
	if !ok {
		return errors.New("Command not found: " + command)
	}
	b := []byte(cmd)
	_, err := p.pipeIn.Write(b)

	return err
}
