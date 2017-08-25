package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"
)

// Player represents the omxplayer
type Player struct {
	api      *apiHandler
	command  *exec.Cmd
	pipeIn   io.WriteCloser
	playlist playlist
	conf     config
}

var commandList = map[string]string{
	"speedIncrease":   "1",
	"speedDecrease":   "2",
	"rewind":          "<",
	"fastForward":     ">",
	"chapterPrevious": "i",
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
func (p *Player) Start(fileName string, position time.Duration) error {
	var err error
	pos := fmt.Sprintf("%02d:%02d:%02d", int(position.Hours()), int(position.Minutes())%60, int(position.Seconds())%60)

	cmd := "omxplayer"

	if p.api.debug {
		cmd = "echo"
	}

	// quit the current video to start the next
	if p.command != nil && p.command.ProcessState == nil {
		if err := p.SendCommand("quit"); err != nil {
			return err
		}
	}

	p.command = exec.Command(cmd, "-b", "-l", pos, path.Join(p.conf.Directory, fileName))
	p.pipeIn, err = p.command.StdinPipe()

	if err != nil {
		return err
	}

	p.command.Stdout = os.Stdout
	err = p.command.Start()

	// wait for the program to exit
	go p.wait()

	return err
}

func (p *Player) wait() {
	// wait for the process to end
	p.command.Wait()
	if p.api.debug {
		log.Println("Process ended")
	}
	// p.playlist.current = nil
}

// SendCommand sends a command to the omxplayer process
func (p *Player) SendCommand(command string) error {
	cmd, ok := commandList[command]
	if !ok {
		return errors.New("Command not found: " + command)
	}

	var err error
	if p.api.debug {
		fmt.Println("cmd:", cmd)
	} else {
		b := []byte(cmd)
		_, err = p.pipeIn.Write(b)
	}

	return err
}

// Handles requets to the player api
func (p *Player) ServeHTTP(w http.ResponseWriter, h *http.Request) {

	if p.api.message.Method == "start" {
		fileName, ok := p.api.message.Arguments["path"]
		if !ok {
			m := &resMessage{Success: false, Message: "No movie name provided."}
			log.Println(m.Message)
			json.NewEncoder(w).Encode(m)
			return
		}

		var position = time.Duration(0)
		if pos, ok := p.api.message.Arguments["position"]; ok {
			p, err := time.ParseDuration(pos)
			if err != nil {
				m := &resMessage{Success: false, Message: "Error converting video position " + err.Error()}
				log.Println(m.Message)
				json.NewEncoder(w).Encode(m)
				return
			}

			position = p
		}

		i := p.playlist.getIndex(fileName)
		if i == -1 {
			m := &resMessage{Success: false, Message: "Trying to play a video that's not in the playlist: " + fileName}
			log.Println(m.Message)
			json.NewEncoder(w).Encode(m)
			return
		}

		err := p.Start(p.playlist.Items[i].Name(), position)
		if err != nil {
			m := &resMessage{Success: false, Message: "Error trying to start video: " + err.Error()}
			log.Println(m.Message)
			json.NewEncoder(w).Encode(m)
			return
		}

		p.playlist.current = p.playlist.Items[i]

		m := &resMessage{Success: true, Message: p.playlist.current.Name()}
		json.NewEncoder(w).Encode(m)
		return
	}

	if p.api.message.Method == "next" {
		err := p.next()
		if err != nil {
			m := &resMessage{Success: false, Message: "Error going to next video: " + err.Error()}
			log.Println(m.Message)
			json.NewEncoder(w).Encode(m)
			return
		}

		m := &resMessage{
			Success: true,
			Message: map[string]interface{}{
				"item": p.playlist.current.Name(),
			}}
		json.NewEncoder(w).Encode(m)
		return
	}

	if p.api.message.Method == "previous" {
		err := p.previous()
		if err != nil {
			m := &resMessage{Success: false, Message: "Error going to previous video: " + err.Error()}
			log.Println(m.Message)
			json.NewEncoder(w).Encode(m)
			return
		}

		m := &resMessage{Success: true, Message: p.playlist.current.Name()}
		json.NewEncoder(w).Encode(m)
		return
	}

	if p.api.message.Method == "sendCommand" {
		cmd, ok := p.api.message.Arguments["command"]
		if !ok {
			m := &resMessage{Success: false, Message: "No command sent."}
			log.Println(m.Message)
			json.NewEncoder(w).Encode(m)
			return
		}

		err := p.SendCommand(cmd)
		if err != nil {
			m := &resMessage{Success: false, Message: "Error trying to execute command: " + err.Error()}
			log.Println(m.Message)
			json.NewEncoder(w).Encode(m)
			return
		}

		m := &resMessage{Success: true, Message: "Command sent and executed"}
		json.NewEncoder(w).Encode(m)
		return
	}

	m := &resMessage{Success: false, Message: "Method not supported"}
	log.Println(m.Message)
	json.NewEncoder(w).Encode(m)
	return
}

// Scan the folder for new files every time the page reloads
func (p *Player) handleControl() http.Handler {
	err := p.playlist.fromFolder(p.conf.Directory)
	if err != nil {
		log.Println("Error tring to read files from directory: ", err)
	}

	tempControl := templateHandler{
		filename: "control.html",
		data: map[string]interface{}{
			"directory": p.conf.Directory,
			"playlist":  p.playlist,
			"error":     err,
		},
	}

	return &tempControl
}

func (p *Player) next() error {
	n := p.playlist.getNext()
	err := p.Start(n.Name(), time.Duration(0))

	if err == nil {
		p.playlist.current = n
	}

	return err
}

func (p *Player) previous() error {
	n := p.playlist.getPrevious()
	err := p.Start(n.Name(), time.Duration(0))

	if err == nil {
		p.playlist.current = n
	}

	return err
}
