package main

import (
	"errors"
	"io/ioutil"
	"path"
)

type playlist struct {
	name  string
	items []string
}

func (p *playlist) fromFolder(folderPath string) error {
	// read files from a certain folder into a playlist
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return errors.New("can't read folder for videos: " + err.Error())
	}

	// filter out all files except for .mp4s
	for _, file := range files {
		if path.Ext(file.Name()) == ".mp4" {
			p.items = append(p.items, file.Name())
		}
	}

	return nil
}
