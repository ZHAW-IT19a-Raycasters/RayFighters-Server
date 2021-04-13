package gamemap

import "testing"

func TestMap(t *testing.T) {
	filePath := "gamemap//testmap_01_small_valid.txt"
	const mapDimention = 5
	/*
		gm_expected := [mapDimention][mapDimention]MapElement{
			{Floor, Floor, Floor, Floor, Floor},
			{Floor, Floor, Floor, Floor, Floor},
			{Floor, Floor, Floor, Floor, Floor},
			{Floor, Floor, Floor, Floor, Floor},
			{Floor, Floor, Floor, Floor, Floor},
		}*/
	gm_gotten, err := MapParser(filePath)

	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}

	for lineIndex := 0; lineIndex < mapDimention; lineIndex += 1 {
		if len(gm_gotten.GMap[lineIndex]) != mapDimention {
			t.Errorf("Error: map dimantion missmatch expected: %d actual: %d", mapDimention, len(gm_gotten.GMap[0]))
			return
		}
	}

}
