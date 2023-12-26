package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

// Saves is the map of all save information.
var Saves = map[string]*SaveInfo{}

// SaveInfo holds metadata about a save.
type SaveInfo struct {
	ID   string   // Save ID, save path is saves/[ID]/
	Name string   // Human-readable
	Mods []string // List of mods used when creating the save
}

// LoadSaveInfo refreshes all saves data.
func LoadSaveInfo() error {
	Saves = map[string]*SaveInfo{}
	files, err := os.ReadDir("saves")
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		p := path.Join("saves", file.Name())
		d, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		si := SaveInfo{}
		if err := json.Unmarshal(d, &si); err != nil {
			return err
		}
		si.ID = strings.TrimRight(file.Name(), ".json")
		if _, found := Saves[si.ID]; found {
			return fmt.Errorf("duplicate save ID %s", si.ID)
		}
		Saves[si.ID] = &si
	}
	return nil
}
