package gamemap

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"
)

var headerLength = 5

// Error codes returned by failures to parse a map.
var (
	ErrMapDimNeg = errors.New("maploader: negative map dimesion")
	ErrMapDimCol = errors.New("maploader: shorter than expected :(")
	ErrMapHead   = errors.New("maploader: wrong header format")
)

// LoadMap loads square game map from the given file path.
// It panics if any problemns are encountered, while reading the file.
func LoadMap(filePath string) ([][]MapElement, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Errorf(err.Error())
		return nil, err
	}
	defer file.Close()
	return parseMap(file)
}

func parseMap(file io.Reader) ([][]MapElement, error) {
	scanner := bufio.NewScanner(file)
	// Skip header
	err := skipLines(scanner, headerLength-1)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	// Read the map dimesion and make sure it is positive
	mapDimesion, err := readMapDimension(scanner)
	if err != nil {
		logger.Errorf(err.Error())
		return nil, err
	}

	// Convert map in the file into in gamemap structure
	locationMap := make([][]MapElement, mapDimesion)
	for rowIndex := 0; rowIndex < mapDimesion; rowIndex += 1 {
		if !scanner.Scan() {
			logger.Error(ErrMapDimCol)
			return nil, ErrMapDimCol
		}
		res, err := parseLine(scanner.Text())
		locationMap[rowIndex] = res
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
	}

	return locationMap, nil
}

// skipLines skips "lineCount" lines.
func skipLines(sc *bufio.Scanner, linesCount int) error {
	for i := 0; i < linesCount; i += 1 {
		if !sc.Scan() {
			logger.Error(ErrMapHead.Error())
			return ErrMapHead
		}
	}
	return nil
}

// readMapDimension reads int from the next line.
func readMapDimension(sc *bufio.Scanner) (int, error) {
	if !sc.Scan() {
		logger.Error(ErrMapDimCol.Error())
		return 0, ErrMapDimCol
	}
	mapDimesion, err := strconv.Atoi(sc.Text())
	if err != nil {
		logger.Error(err.Error())
		return 0, err
	}
	if mapDimesion < 0 {
		logger.Error(ErrMapDimNeg)
		return 0, ErrMapDimNeg
	}
	return mapDimesion, nil
}

// parseLine converts string into slice of MapElements.
func parseLine(line string) ([]MapElement, error) {
	slice := make([]MapElement, len(line))
	for rowIndex, c := range line {
		element, err := convert(c)
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
		slice[rowIndex] = element
	}
	return slice, nil
}
