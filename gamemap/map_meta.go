package gamemap

type MapMeta struct {
	id				uint8
	name, filePath 	string
}

func NewMapMeta(id uint8, name, filePath string) MapMeta {
	return MapMeta{
		id:       id,
		name:     name,
		filePath: filePath,
	}
}

func (mm MapMeta) GetId() {
	
}

func (mm MapMeta) GetName() {

}

func (mm MapMeta) GetFilePath() {

}
