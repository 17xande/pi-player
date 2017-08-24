package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
)

type playlist struct {
	Name    string
	Items   []os.FileInfo
	current os.FileInfo
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
			p.Items = append(p.Items, file)
		}
	}

	return nil
}

func (p *playlist) getIndex(fileName string) int {
	for i, item := range p.Items {
		if item.Name() == fileName {
			return i
		}
	}

	return -1
}

func (p *playlist) getNext() os.FileInfo {
	i := p.getIndex(p.current.Name())
	if i+1 > len(p.Items) {
		return p.Items[0]
	}

	return p.Items[i]
}

func (p *playlist) getPrevious() os.FileInfo {
	i := p.getIndex(p.current.Name())
	if i-1 < 0 {
		return p.Items[len(p.Items)-1]
	}

	return p.Items[i]
}
