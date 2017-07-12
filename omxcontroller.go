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
	"increase-speed":   "1",
	"decrease-speed":   "2",
	"rewind":           "<",
	"fast-forward":     ">",
	"previous-chapter": "i",
	"next-chapter":     "o",
	"exit":             "q",
	"quit":             "q",
	"pause-resume":     "p",
	"decrease-volume":  "-",
	"increase-volume":  "+",
	"seek-back-30":     "\x1b[D",
	"seek-forward-30":  "\x1b[C",
	"seek-back-600":    "\x1b[B",
	"seek-forward-600": "\x1b[A",
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
