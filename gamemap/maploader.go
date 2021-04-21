package gamemap

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

var headerLength = 5

// Error codes returned by failures to parse a map.
var (
	ErrShoF   = errors.New("maploader: internal error")
	ErrMapDim = errors.New("maploader: invalid map dimension")
)

func check(e error) {
	if e != nil {
		logger.Errorf(e.Error())
		panic(e)
	}
}

// LoadMap loads square game map from the given file path.
// It panics if any problemns are encountered, while reading the file.
func LoadMap(filePath string) ([][]MapElement, error) {
	file, err := os.Open(filePath)
	check(err)
	defer file.Close()
	return parseMap(file)
}

func parseMap(file io.Reader) ([][]MapElement, error) {
	scanner := bufio.NewScanner(file)
	err := skipLines(scanner, headerLength-1)
	if err != nil {
		logger.Errorf("unexpected error: %s", err.Error())
		return [][]MapElement{{}}, fmt.Errorf("unexpected error: %s", err.Error())
	}
	mapDimesion, err := readMapDimension(scanner)
	if err != nil {
		logger.Errorf("unexpected error: %s", err.Error())
		return [][]MapElement{{}}, fmt.Errorf("unexpected error: %s", err.Error())
	}

	// go row by row read symbols to the gamemap
	locationMap := make([][]MapElement, mapDimesion)
	for rowIndex := 0; rowIndex < mapDimesion; rowIndex += 1 {
		if !scanner.Scan() {
			logger.Errorf("unexpected file end. expected length: %d, actual length:", mapDimesion, rowIndex)
			return [][]MapElement{{}}, fmt.Errorf("unexpected file end. expected length: %d, actual length:%d", mapDimesion, rowIndex)
		}
		res, err := parseLine(scanner.Text())
		locationMap[rowIndex] = res
		if err != nil {
			logger.Errorf("unexpected error: %s", err.Error())
			return [][]MapElement{{}}, fmt.Errorf("unexpected error: %s", err.Error())
		}
	}

	return locationMap, nil
}

// Skips "lineCount" lines.
// NOTE: To get content of the line n+1: use sc.Scan() first.
// Throws error if:
// - file is too short for given "lineCount"
func skipLines(sc *bufio.Scanner, linesCount int) error {
	for i := 0; i < linesCount; i += 1 {
		if !sc.Scan() {
			return fmt.Errorf("can not skip: file too short")
		}
	}
	return nil
}

// readMapDimension reads int from the next line.
// Throws error if:
// - there no lines to read
// - there is int conversion error
// - gotten number is negative
func readMapDimension(sc *bufio.Scanner) (int, error) {
	if !sc.Scan() {
		return 0, fmt.Errorf("unexpected end of file")
	}
	mapDimesion, err := strconv.Atoi(sc.Text())
	if err != nil {
		return 0, err
	}
	if mapDimesion < 0 {
		return 0, ErrMapDim
	}
	return mapDimesion, nil
}

// parseLine converts string into slice of MapElments.
// Throws error if:
// - utils.LocationEelment() fails
func parseLine(line string) ([]MapElement, error) {
	slice := make([]MapElement, len(line))
	for rowIndex, c := range line {
		element, err := convert(c)
		if err != nil {
			return []MapElement{}, err
		}
		slice[rowIndex] = element
	}
	return slice, nil
}
