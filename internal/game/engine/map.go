package engine

import (
	"encoding/json"
	"fmt"
	"os"
)

const DefaultMapWidth = 100
const DefaultMapHeight = 100

type MapData struct {
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Tiles  [][]int `json:"tiles"`
}

func LoadMapData(path string) (MapData, error) {
	file, err := os.Open(path)
	if err != nil {
		return MapData{}, err
	}
	defer file.Close()

	var data MapData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return MapData{}, err
	}

	if err := validateMapData(data); err != nil {
		return MapData{}, err
	}

	return data, nil
}

func DefaultMapData(width, height int) MapData {
	return MapData{
		Width:  width,
		Height: height,
		Tiles:  buildMapTiles(width, height),
	}
}

func validateMapData(data MapData) error {
	if data.Width <= 0 || data.Height <= 0 {
		return fmt.Errorf("invalid map size")
	}

	if len(data.Tiles) != data.Height {
		return fmt.Errorf("invalid map rows")
	}

	for y := 0; y < data.Height; y += 1 {
		if len(data.Tiles[y]) != data.Width {
			return fmt.Errorf("invalid map columns")
		}
	}

	return nil
}

func buildMapTiles(width, height int) [][]int {
	data := make([][]int, 0, height)

	for y := 0; y < height; y += 1 {
		row := make([]int, 0, width)
		for x := 0; x < width; x += 1 {
			isBorder := x == 0 || y == 0 || x == width-1 || y == height-1
			tileIndex := 0

			if isBorder {
				tileIndex = 2
			} else if (x+y)%7 == 0 {
				tileIndex = 1
			}

			row = append(row, tileIndex)
		}
		data = append(data, row)
	}

	return data
}
