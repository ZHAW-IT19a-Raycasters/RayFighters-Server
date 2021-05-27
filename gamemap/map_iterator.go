package gamemap

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

// MapIteratorManager is the parent instance of all
// the iterators. It is responsible for loading
// parsing and iterating the maps around.
type MapIteratorManager struct {
	// filePath is the path to the JSON file
	filePath string
	// maps is the channel that contains all the
	// maps at any time.
	maps chan *MapMeta
}

// NewMapIteratorManager will create a new instance of
// the MapIteratorManager.
func NewMapIteratorManager(filePath string) (MapIteratorManager, error) {
	ml, err := readJSON(filePath)
	if err != nil {
		return MapIteratorManager{}, err
	}

	mapChannel := make(chan *MapMeta, len(ml.maps))
	for _, element := range ml.maps {
		tmp := element.mapMeta()
		mapChannel <- &tmp
	}
	return MapIteratorManager{
		filePath: filePath,
		maps:     mapChannel,
	}, nil
}

// GetFilePath will return the file path to the index.json
// file from which the maps where loaded.
func (mim *MapIteratorManager) GetFilePath() string {
	return mim.filePath
}

// GetIterator will create a child iterator, to be used by
// an individual hub.
func (mim *MapIteratorManager) GetIterator() *MapIterator {
	return newMapIterator(mim.maps)
}

// MapIterator is the child instance of the
// MapIteratorManager. It itself is responsible
// for returning the map to the channel after
// reading it.
type MapIterator struct {
	// maps is the channel with all the maps
	maps chan *MapMeta
}

// newMapIterator will create a new iterator,
func newMapIterator(channel chan *MapMeta) *MapIterator {
	return &MapIterator{
		maps: channel,
	}
}

// GetMap will return the next map to be loaded
func (mi *MapIterator) GetMap() *MapMeta {
	m := <-mi.maps
	mi.maps <- m
	return m
}

// readJSON will load the index.json file from
// the hard drive and parse the content directly.
func readJSON(filePath string) (mapList, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return mapList{}, err
	}

	return parseJSON(content, filePath)
}

// IsValid is a small helper function that
// checks if the given string is a valid
// index.json file or not.
func IsValid(jsonPath string) error {
	_, err := NewMapIteratorManager(jsonPath)
	return err
}

// parseJSON is a helper function responsible
// for parsing the json data containing
// all the map information.
func parseJSON(jsonContent []byte, jsonPath string) (mapList, error) {
	var maps []mapMetaJSON
	err := json.Unmarshal(jsonContent, &maps)

	if err != nil {
		return mapList{}, err
	}

	ml := mapList{
		maps: maps,
	}

	// get parent path
	absPath, err := filepath.Abs(jsonPath)
	if err != nil {
		return mapList{}, err
	}
	// fix path
	parent := filepath.Dir(absPath)

	for i, m := range ml.maps {
		m.FilePath = filepath.Join(parent, m.FilePath)
		ml.maps[i] = m
	}

	return ml, nil
}
