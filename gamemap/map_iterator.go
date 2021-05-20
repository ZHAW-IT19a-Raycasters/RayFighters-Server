package gamemap

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

// TODO: COMMENTS!!! --> Dali

type MapIteratorManager struct {
	// filePath is the path to the JSON file
	filePath string
	maps     chan *MapMeta
}

func (mim *MapIteratorManager) GetFilePath() string {
	return mim.filePath
}

func newMapIterator(channel chan *MapMeta) *MapIterator {
	return &MapIterator{
		maps: channel,
	}
}

func (mim *MapIteratorManager) GetIterator() *MapIterator {
	return newMapIterator(mim.maps)
}

func NewMapIteratorManager(filePath string) (MapIteratorManager, error) {
	// TODO: Read Content
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

type MapIterator struct {
	maps chan *MapMeta
}

func (mi *MapIterator) GetMap() *MapMeta {
	m := <-mi.maps
	mi.maps <- m
	return m
}

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

func readJSON(filePath string) (mapList, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return mapList{}, err
	}

	return parseJSON(content, filePath)
}

func IsValid(jsonPath string) error {
	_, err := NewMapIteratorManager(jsonPath)
	return err
}
