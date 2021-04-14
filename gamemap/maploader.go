package gamemap

import (
	"bufio"
	"os"
	"strconv"
)

type GameMap struct {
	GMap [][]MapElement //slice
}

func (gm GameMap) GetMapElement(coord_x int, coord_y int) MapElement {
	return gm.GMap[coord_x][coord_y]
}

// Declaring enum
type MapElement int

// 0 = error
const (
	Floor MapElement = iota + 1
	Wall
	MedkitSpawn
	WreaponSpawn
)

//switch(static) vs. map (dynamic)
func runeToMapElement(r rune) MapElement {
	switch r {
	case '0':
		return Floor
	case '+':
		return MedkitSpawn
	case '*':
		return WreaponSpawn
	default:
		return Wall
	}
}

const numberOfLinesInMapHeader int = 5

func MapParser(filePath string) (GameMap, error) {
	var gameMap GameMap

	file, err := os.Open(filePath)
	if err != nil {
		return GameMap{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var mapDimesion int

	// skip first 'numberOfLinesInMapHeader' of header
	for i := 0; i < numberOfLinesInMapHeader; i += 1 {
		scanner.Scan()
	}
	mapDimesion, _ = strconv.Atoi(scanner.Text())
	gameMap.GMap = make([][]MapElement, 0, mapDimesion)

	for rowIndex := 0; rowIndex < mapDimesion; rowIndex += 1 {
		gameMap.GMap = append(gameMap.GMap, make([]MapElement, 0, mapDimesion))

		scanner.Scan()

		line := scanner.Text()
		parseLine(line, &gameMap.GMap[rowIndex])
	}

	return gameMap, nil
}

func parseLine(line string, gmRow *[]MapElement) {
	for _, c := range line {
		*gmRow = append(*gmRow, runeToMapElement(c))
	}
}
