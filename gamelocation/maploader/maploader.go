package maploader

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.zhaw.ch/raycasters/server/gamelocation/utils"
	"github.zhaw.ch/raycasters/server/log"
)

var mapHeaderLength int = 5
var logger = log.Log()

/*
const (
	// Define map-header length
	mapHeaderLength int = 5
	maxMapDimesion  int = 999
)
*/

// Loads file from given path.
// Throws error if:
// - could open or read file under given path
// - SkipLines() fails
// - ReadDimesion() fails
// - file length was shorter than expected
// - ParseLine() fails
func LoadMap(filePath string) (*[][]utils.LocationElement, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return &[][]utils.LocationElement{{}}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	err = SkipLines(scanner, mapHeaderLength)
	if err != nil {
		logger.Errorf("unexpected error: %s", err.Error())
		return &[][]utils.LocationElement{{}}, fmt.Errorf("unexpected error: %s", err.Error())
	}
	mapDimesion, err := ReadMapDimension(scanner)
	if err != nil {
		logger.Errorf("unexpected error: %s", err.Error())
		return &[][]utils.LocationElement{{}}, fmt.Errorf("unexpected error: %s", err.Error())
	}

	// go row by row read symbols to the gamemap
	locationMap := make([][]utils.LocationElement, mapDimesion)
	for rowIndex := 0; rowIndex < mapDimesion; rowIndex += 1 {
		if scanner.Scan() {
			logger.Errorf("unexpected file end. expected length: %d, actual length:", mapDimesion, rowIndex)
			return &[][]utils.LocationElement{{}}, fmt.Errorf("unexpected file end. expected length: %d, actual length:%d", mapDimesion, rowIndex)
		}
		res, err := ParseLine(scanner.Text())
		locationMap[rowIndex] = *res
		if err != nil {
			logger.Errorf("unexpected error: %s", err.Error())
			return &[][]utils.LocationElement{{}}, fmt.Errorf("unexpected error: %s", err.Error())
		}
	}

	return &locationMap, nil
}

// Skips "lineCount" lines.
// NOTE: To get content of the line n+1: use sc.Scan() first.
// Throws error if:
// - file is too short for given "lineCount"
func SkipLines(sc *bufio.Scanner, linesCount int) error {
	for i := 0; i < linesCount; i += 1 {
		if !sc.Scan() {
			return fmt.Errorf("can not skip: file too short")
		}
	}
	return nil
}

// Reads dimension from the next line
// Return integer.
// Throws error if:
// - there no lines to read
// - no element could be converted into integer
// - dimension <0
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
// Throws error if:
// - utils.LocationElment() fails
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
