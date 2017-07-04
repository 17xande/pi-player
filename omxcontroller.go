package main

import (
	"io"
	"os"
	"os/exec"
)

// Player represents the omxplayer
type Player struct {
	command *exec.Cmd
	pipeIn  io.WriteCloser
}

var commandList = map[string]string{
	"playppause": "p",
	"up":         "\x1b[A",
	"down":       "\x1b[B",
	"left":       "\x1b[D",
	"right":      "\x1b[C",
}

// Start starts the player
func (p Player) Start(path string) error {
	var err error
	p.command = exec.Command("omxplayer", "-b", path)
	p.pipeIn, err = p.command.StdinPipe()
	if err != nil {
		return err
	}

	p.command.Stdout = os.Stdout
	err = p.command.Start()
	return err
}

// SendCommand sends a command to the omxplayer process
func (p Player) SendCommand(command string) error {
	_, err := p.pipeIn.Write([]byte(commandList[command]))

	return err
}
