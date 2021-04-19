package maploader

import (
	"bufio"
	"fmt"
	"strconv"

	"github.zhaw.ch/raycasters/server/gamelocation/utils"
)

//var logger = log.Log()
/*
const (
	// Define map-header length
	mapHeaderLength int = 5
	maxMapDimesion  int = 999
)
*/
/*
// Loads file from given path.
// *PathError, if there were any problems while opening and reading the file at the given location.
func LoadMap(filePath string) ([][]utils.LocationElement, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return [][]utils.LocationElement{{}}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	SkipLines(scanner, mapHeaderLength)
	mapDimesion, _ := ReadMapDimension(scanner)
	// check correctness of dimensions
	if mapDimesion < 0 {
		logger.Errorf("MAPLOADER ERROR: negative dimesion.")
		return [][]utils.LocationElement{{}}, fmt.Errorf("MAPLOADER ERROR: negative dimesion.")
	}
	// go row by row read symbols to the gamemap
	for rowIndex := 0; rowIndex < mapDimesion; rowIndex += 1 {
		gameMap := append(gameMap.GMap, make([]MapElement, 0, mapDimesion))

		if scanner.Scan() {
			logger.Errorf("MAPLOADER ERROR: unexpected file end")
			return [][]utils.LocationElement{{}}, fmt.Errorf("ERROR: unexpected end of file")
		}

		line := scanner.Text()
		parseLine(line, &gameMap.GMap[rowIndex])
		if len(line) != mapDimesion {
			logger.Errorf("MAPLOADER ERROR: unexpected file end. #lines(Act.: %d, Exp: %d)", mapDimesion, rowIndex+1)
			return [][]utils.LocationElement{{}}, fmt.Errorf("ERROR: unexpected end of file")
		}
	}
	location := gamelocation.NewLocation(mapDimesion)

	return location, nil
}*/
// Skips n lines.
// To get content of the line n+1: use sc.Scan() first.
func SkipLines(sc *bufio.Scanner, linesCount int) error {
	for i := 0; i < linesCount; i += 1 {
		if !sc.Scan() {
			return fmt.Errorf("can not skip: file too short")
		}
	}
	return nil
}

// Reads dimension from the next line
// Throws error if dimension <0
func ReadMapDimension(sc *bufio.Scanner) (int, error) {
	if !sc.Scan() {
		return 0, fmt.Errorf("unexpected end of file")
	}
	mapDimesion, err := strconv.Atoi(sc.Text())
	if err != nil {
		return 0, err
	}
	if mapDimesion < 0 {
		return 0, fmt.Errorf("invalid map dimension")
	}
	return mapDimesion, nil
}

// Converts string into slice of LocationElement.
func ParseLine(line string) (*([]utils.LocationElement), error) {
	slice := make([]utils.LocationElement, len(line))
	for rowIndex, c := range line {
		element, err := utils.GetLocationElement(c)
		if err != nil {
			return &([]utils.LocationElement{}), err
		}
		slice[rowIndex] = element
	}
	return &slice, nil
}
