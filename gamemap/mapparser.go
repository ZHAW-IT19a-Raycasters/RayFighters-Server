package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

type Coordinates struct {
	x, y int
}
type ObjectType int

//0 is error
const (
	Floor = iota + 1
	Wall
	MedkitSpawn
	WreaponSpawn
)

// if not found
var runeToObjectType = map[rune]ObjectType{
	rune('0'): Floor,
	rune('1'): Wall,
	rune('+'): MedkitSpawn,
	rune('*'): WreaponSpawn,
}

const numberOfLinesInMapHeader int = 5

var GameMap map[Coordinates]ObjectType

func main() {
	//fmt.Println(byte("0"))
	//fmt.Println(runeToObjectType[rune('+')])
	GameMap = make(map[Coordinates]ObjectType)
	//file, err := os.Open("gamemap/testmap_01_valid.txt")
	file, err := os.Open("gamemap/testmap_02_valid_specialsymbols.txt")
	//file, err := os.Open("gamemap/testmap_04_illegalsymbol.txt")

	if err != nil {
		fmt.Println("%s", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var mapDimesion int
	count := 0

	row := 0
	for scanner.Scan() {
		count++
		//read map dimention
		if count == numberOfLinesInMapHeader {
			mapDimesion, err = strconv.Atoi(scanner.Text())
			if err != nil {
				fmt.Printf("%s: %s\n", "Invalid map format: expected int map dimesion", err)
				return
			}
		}
		if count > numberOfLinesInMapHeader {
			row = count - numberOfLinesInMapHeader - 1
			line := scanner.Text()

			//check if map has defined columns of rows
			if len(line) != mapDimesion {
				fmt.Printf("%s. Expected: %d Is: %d\n", "Map missformat: row length missmatch", mapDimesion, len(line))
				return
			}

			// fill the row
			column := 0
			for _, c := range line { //??what for is the 2 argument in range
				var obj ObjectType = runeToObjectType[c]
				if obj == 0 {
					fmt.Printf("Unsupported symbol: %s", string(c))
					return
				}
				fmt.Printf("rune: %d obj: %d\n", c, obj)
				GameMap[Coordinates{row, column}] = obj
				column++
			}
		}

	}
	//check if map has defined number of rows > no need because is a quadr
	if mapDimesion != (count - numberOfLinesInMapHeader) {
		fmt.Printf("%s: \n", "Invalid map format: lines count missmatch")
		return
	}

	fmt.Println("map dimension is:", mapDimesion)
	fmt.Println(GameMap)
	fmt.Println(GameMap[Coordinates{0, 1}])
	//fmt.Println(runeToObjectType[rune('.')])// for non existing symbol map returns 0
}
