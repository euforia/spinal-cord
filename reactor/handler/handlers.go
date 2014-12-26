package handler

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
)

type Handler struct {
	Path string
	Data []byte          `json:"data"`
	Sha1 [sha1.Size]byte `json:"sha1"`
}

func NewHandler(path string, data []byte) (*Handler, error) {
	sha1sum := sha1.Sum(data)
	return &Handler{path, data, sha1sum}, nil
}

func (h *Handler) Sha1String() string {
	return fmt.Sprintf("%x", h.Sha1)
}

func (h *Handler) WriteHandlerFile(dirpath string, perms os.FileMode) error {
	abspath := fmt.Sprintf("%s/%s", dirpath, h.Path)
	//fmt.Printf("%s/handlers/%s\n", dirpath, h.Path)
	return ioutil.WriteFile(abspath, h.Data, perms)
}

func (h *Handler) Remove() error {
	return os.Remove(h.Path)
}

type EventHandler struct {
	Path     string `json:"path"`
	FullPath string `json:"fullpath"`
	Name     string `json:"name"`
}

func (eh *EventHandler) Handler() (Handler, error) {
	contents, err := ioutil.ReadFile(eh.FullPath)
	if err != nil {
		var empty [20]byte
		return Handler{eh.Path, contents, empty}, err
	}
	sha1sum := sha1.Sum(contents)
	//"%x", sha1sum
	return Handler{eh.Path, contents, sha1sum}, nil
}

func GetHandlerFromFile(path string) (*Handler, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	sha1sum := sha1.Sum(data)
	return &Handler{path, data, sha1sum}, nil
}
