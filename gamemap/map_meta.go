package gamemap

import "path/filepath"

// mapList is a struct of mapMetaJSON Array
type mapList struct {
	maps []mapMetaJSON
}

// mapMetaJSON saves the map from the JSON file
type mapMetaJSON struct {
	Id             uint8
	Name, FilePath string
}

// mapMeta creates a MapMeta from a mapMetaJSON element
func (m *mapMetaJSON) mapMeta() MapMeta {
	return MapMeta{
		id:       m.Id,
		name:     m.Name,
		filePath: m.FilePath,
	}
}

// MapMeta saves the internal meta information of a map
type MapMeta struct {
	id             uint8
	name, filePath string
}

// getId returns the id of a map
func (mm MapMeta) GetId() uint8 {
	return mm.id
}

// getName returns the name of a map
func (mm MapMeta) GetName() string {
	return mm.name
}

// getFilePath returns the file path of a map
func (mm MapMeta) GetFilePath() string {
	return mm.filePath
}

// Loadfile generates a map2d struct from the MapMeta
// it returns the map2d element and a error
func (mm MapMeta) LoadFile() (Map2d, error) {
	return LoadMap2d(filepath.Join(mm.GetFilePath(), mm.GetName()))
}
