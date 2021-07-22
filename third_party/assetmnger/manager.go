package assetmnger

import (
	"github.com/gobuffalo/packr/v2"
)

const (
	DIRPATH = "../../assets"
)

type Manager struct {
	configBox *packr.Box
}

func (m *Manager) Get(fileName string) []byte {
	byteValue, err := m.configBox.Find(fileName)
	if err != nil {
		return nil
	}
	return byteValue
}

func NewManager() *Manager {
	var m = &Manager{}
	box := packr.New("configBox", DIRPATH)
	m.configBox = box
	return m
}
