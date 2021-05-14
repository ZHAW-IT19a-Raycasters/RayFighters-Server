package room

import (
	"encoding/json"
	"github.zhaw.ch/raycasters/server/gamemap"
	"io/ioutil"
)

// TODO: COMMENTS!!! --> Dali

type mapList struct {
	maps []gamemap.MapMeta
}

type MapIteratorManager struct {
	// filePath is the path to the JSON file
	filePath string
	maps chan gamemap.MapMeta
}

func newMapIterator(channel chan gamemap.MapMeta) MapIterator {
	return MapIterator{
		maps: channel,
	}
}

func (mim *MapIteratorManager) GetIterator() MapIterator{
	return newMapIterator(mim.maps)
}

type MapIterator struct {
	maps chan gamemap.MapMeta
}

func NewMapIteratorManager(filePath string) (MapIteratorManager, error) {
	// TODO: Read Content
	ml, err := readJSON(filePath)
	if err != nil {
		return MapIteratorManager{}, err
	}

	mapChannel := make(chan gamemap.MapMeta, len(ml.maps))
	for _, element := range ml.maps {
		mapChannel <- element
	}

	return MapIteratorManager{
		filePath: filePath,
		maps: mapChannel,
	}, nil
}

func (mi *MapIterator) GetMap() gamemap.MapMeta {
	m := <-mi.maps
	mi.maps <- m
	return m
}

func parseJSON(jsonContent []byte) (mapList, error) {
	ml := mapList{}
	err := json.Unmarshal(jsonContent, &ml)
	return ml, err
}

func readJSON(jsonContent string) (mapList, error) {
	content, err := ioutil.ReadFile(jsonContent)
	if err != nil {
		return mapList{}, err
	}

	return parseJSON(content)
}
