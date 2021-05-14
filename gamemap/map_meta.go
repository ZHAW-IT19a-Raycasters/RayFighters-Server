package gamemap

type MapMeta struct {
	id				uint8
	name, filePath 	string
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
	return LoadMap2d(mm.GetFilePath())
}
