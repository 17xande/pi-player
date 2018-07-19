package piplayer

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
)

// Playlist stores the media items that can be played
type Playlist struct {
	Name    string
	Items   []Item
	Current *Item
}

// NewPlaylist creates a new playlist with media in the designated folder.
func NewPlaylist(dir string) (Playlist, error) {
	pl := Playlist{Name: dir}
	err := pl.fromFolder(dir)
	return pl, err
}

// Handles requests to the playlist api
func (p *Playlist) handleAPI(api *APIHandler, w http.ResponseWriter, h *http.Request) {
	var m *resMessage

	switch api.message.Method {
	case "getCurrent":
		if p.Current != nil {
			m = &resMessage{
				Success: true,
				Event:   "current",
				Message: p.Current.Name(),
			}
		} else {
			m = &resMessage{
				Success: true,
				Event:   "noCurrent",
			}
		}
	case "setCurrent":
		if api.message.Arguments == nil || len(api.message.Arguments) == 0 {
			m = &resMessage{
				Success: false,
				Event:   "noArgumentSupplied",
			}
			break
		}

		index, err := strconv.Atoi(api.message.Arguments["index"])

		if err != nil || index < 0 || index >= len(p.Items) {
			m = &resMessage{
				Success: false,
				Event:   "argumentInvalid",
			}
			break
		}

		p.Current = &p.Items[index]

		m = &resMessage{
			Success: true,
			Event:   "setCurrent",
			Message: p.Current,
		}

		if api.debug {
			log.Println("set current item index to:", index)
		}
	case "getItems":
		if err := p.fromFolder(p.Name); err != nil {
			log.Printf("Api call failed. Can't get items from folder %s\n%v", p.Name, err)
		}

		m = &resMessage{
			Success: true,
			Event:   "items",
			Message: p.itemsString(),
		}
	default:
		log.Printf("API call unsupported. Ignoring:\n%v\n", api.message)
	}

	json.NewEncoder(w).Encode(m)
}

func (p *Playlist) fromFolder(folderPath string) error {
	// remove all items from the current playlist if there are any
	p.Items = []Item{}

	// read files from a certain folder into a playlist
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return errors.New("can't read folder for items: " + err.Error())
	}

	// filter out all files except for supported ones
	for _, file := range files {
		e := path.Ext(file.Name())
		if e == ".mp4" {
			p.Items = append(p.Items, Item{Visual: file, Type: "video"})
		} else if e == ".jpg" || e == ".jpeg" || e == ".png" {
			p.Items = append(p.Items, Item{Visual: file, Type: "image"})
		} else if e == ".html" {
			p.Items = append(p.Items, Item{Visual: file, Type: "browser"})
		}
	}

	// scan for .mp3 files to see if any need to be attached to image files
	for _, file := range files {
		e := path.Ext(file.Name())
		if e == ".mp3" {
			audioNoExt := file.Name()[0 : len(file.Name())-len(e)]
			for i, item := range p.Items {
				visual := item.Visual.Name()
				visualNoExt := visual[0 : len(visual)-len(path.Ext(visual))]
				if audioNoExt == visualNoExt {
					p.Items[i].Audio = file
					break
				}
			}
		}
	}

	return nil
}

func (p *Playlist) getIndex(fileName string) int {
	for i, item := range p.Items {
		if item.Name() == fileName {
			return i
		}
	}

	return -1
}

func (p *Playlist) getNext() (*Item, error) {
	if p.Current == nil {
		return nil, errors.New("no current item, can't get next")
	}

	i := p.getIndex(p.Current.Name())
	if i == -1 {
		return nil, errors.New("can't find index of this item: " + p.Current.Name())
	}
	if i+1 > len(p.Items)-1 {
		return &p.Items[0], nil
	}

	return &p.Items[i+1], nil
}

func (p *Playlist) getPrevious() (*Item, error) {
	if p.Current == nil {
		return nil, errors.New("no current item, can't get previous")
	}

	i := p.getIndex(p.Current.Name())
	if i == -1 {
		return nil, errors.New("can't find index of this item: " + p.Current.Name())
	}
	if i-1 < 0 {
		return &p.Items[len(p.Items)-1], nil
	}

	return &p.Items[i-1], nil
}

func (p *Playlist) itemsString() []ItemString {
	var res []ItemString

	for _, item := range p.Items {
		res = append(res, item.String())
	}

	return res
}
