package gamemap

import (
	"encoding/json"
	"io/ioutil"
)

// TODO: COMMENTS!!! --> Dali

type mapList struct {
	maps []MapMeta
}

type MapIteratorManager struct {
	// filePath is the path to the JSON file
	filePath string
	maps chan MapMeta
}

func newMapIterator(channel chan MapMeta) MapIterator {
	return MapIterator{
		maps: channel,
	}
}

func (mim *MapIteratorManager) GetIterator() MapIterator{
	return newMapIterator(mim.maps)
}

type MapIterator struct {
	maps chan MapMeta
}

func NewMapIteratorManager(filePath string) (MapIteratorManager, error) {
	// TODO: Read Content
	ml, err := readJSON(filePath)
	if err != nil {
		return MapIteratorManager{}, err
	}

	mapChannel := make(chan MapMeta, len(ml.maps))
	for _, element := range ml.maps {
		mapChannel <- element
	}

	return MapIteratorManager{
		filePath: filePath,
		maps: mapChannel,
	}, nil
}

func (mi *MapIterator) GetMap() MapMeta {
	m := <-mi.maps
	mi.maps <- m
	return m
}

func parseJSON(jsonContent []byte) (mapList, error) {
	ml := mapList{
		maps: make([]MapMeta, 0),
	}
	err := json.Unmarshal(jsonContent, &ml.maps)
	return ml, err
}

func readJSON(jsonContent string) (mapList, error) {
	content, err := ioutil.ReadFile(jsonContent)
	if err != nil {
		return mapList{}, err
	}

	return parseJSON(content)
}

func IsValid(jsonPath string) error {
	_, err := NewMapIteratorManager(jsonPath)
	return err
}
