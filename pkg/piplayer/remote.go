package piplayer

import (
	"fmt"
	"log"

	"github.com/17xande/keylogger"
)

type remote struct {
	Name    string
	Vendor  uint16
	Product uint16
}

var directions = []string{"UP", "DOWN", "HOLD"}

func remoteRead(p *Player, cie chan keylogger.InputEvent) {
	var ie keylogger.InputEvent

	if p.api.debug {
		log.Println("starting remote read for this device")
	}

	send := p.ConnViewer.getChanSend()

	for {
		ie = <-cie
		// ignore events that are not EV_KEY events that are KEY_DOWN presses
		if ie.Type != keylogger.EventTypes["EV_KEY"] || directions[ie.Value] != "DOWN" {
			continue
		}

		key := ie.KeyString()
		fmt.Println(ie.Code, ie.Value, ie.Type)

		if p.api.debug {
			log.Println("Key:", key, directions[ie.Value])
			log.Println("Sending keypress to page and nothing else.")
		}

		msg := wsMessage{
			Component: "remote",
			Arguments: map[string]string{"keyString": key},
			Event:     "keyDown",
		}
		send <- msg
	}
}

func (p *Player) handleRemote(e keylogger.InputEvent) {

}
