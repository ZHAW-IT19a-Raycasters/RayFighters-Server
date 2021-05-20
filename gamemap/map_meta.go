package gamemap

import "path/filepath"

type mapList struct {
	maps []mapMetaJSON
}

type mapMetaJSON struct {
	Id             uint8
	Name, FilePath string
}

func (m *mapMetaJSON) mapMeta() MapMeta {
	return MapMeta{
		id:       m.Id,
		name:     m.Name,
		filePath: m.FilePath,
	}
}

type MapMeta struct {
	id             uint8
	name, filePath string
}

func (mm MapMeta) GetId() uint8 {
	return mm.id
}

func (mm MapMeta) GetName() string {
	return mm.name
}

func (mm MapMeta) GetFilePath() string {
	return mm.filePath
}

func (mm MapMeta) LoadFile() (Map2d, error) {
	return LoadMap2d(filepath.Join(mm.GetFilePath(), mm.GetName()))
}
