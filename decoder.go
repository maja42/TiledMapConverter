package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type TileMap struct {
	Width       int    `xml:"width,attr"`
	Height      int    `xml:"height,attr"`
	Version     string `xml:"version,attr"`
	Orientation string `xml:"orientation,attr"`
	Renderorder string `xml:"renderorder,attr"`
	Tilewidth   int    `xml:"tilewidth,attr"`
	Tileheight  int    `xml:"tileheight,attr"`

	Tilesets []TileSet      `xml:"tileset"`
	Layers   []TileMapLayer `xml:"layer"`
}

const (
	FlippedHorizontallyTiledFlag uint32 = 0x80000000
	FlippedVerticallyTiledFlag   uint32 = 0x40000000
	FlippedDiagonallyTiledFlag   uint32 = 0x20000000
)

type TileSetType uint8

const (
	ENVIRONMENT_TILESET TileSetType = 0
	DECORATION_TILESET  TileSetType = 1
	SPAWN_TILESET       TileSetType = 2
)

type TileSet struct {
	Type       TileSetType `xml:"-"`
	FirstGid   uint32      `xml:"firstgid,attr"`
	Name       string      `xml:"name,attr"`
	TileWidth  int         `xml:"tilewidth,attr"`
	TileHeight int         `xml:"tileheight,attr"`
	TileCount  uint32      `xml:"tilecount,attr"`
	Columns    int         `xml:"columns,attr"`
}

type TileMapLayer struct {
	Name    string `xml:"name,attr"`
	RawData string `xml:"data"`
	Tiles   []Tile `xml:"-"`
}

type Tile struct {
	Index   uint32
	Flags   uint8
	TileSet *TileSet
}

const FIRST_DIAGONAL_TILE_TYPE uint32 = 6*8 + 1

type TileType uint8

const (
	COMPLETELY_ACCESSIBLE TileType = 0
	COMPLETELY_SOLID      TileType = 1
	SOLID_AT_UPPER_LEFT   TileType = 2
	SOLID_AT_UPPER_RIGHT  TileType = 3
	SOLID_AT_LOWER_LEFT   TileType = 4
	SOLID_AT_LOWER_RIGHT  TileType = 5
)

type Orientation uint8

const (
	LEFT      Orientation = 0
	RIGHT     Orientation = 1
	UP        Orientation = 2
	DOWN      Orientation = 3
	UPLEFT    Orientation = 4
	UPRIGHT   Orientation = 5
	DOWNLEFT  Orientation = 6
	DOWNRIGHT Orientation = 7
)

func IsOrientationDiagonal(orientation Orientation) bool {
	return uint8(orientation) >= uint8(UPLEFT)
}

func GetInvertedOrientation(orientation Orientation) Orientation {
	switch orientation {
	case LEFT:
		return RIGHT
	case RIGHT:
		return LEFT
	case UP:
		return DOWN
	case DOWN:
		return UP
	case UPLEFT:
		return DOWNRIGHT
	case UPRIGHT:
		return DOWNLEFT
	case DOWNLEFT:
		return UPRIGHT
	case DOWNRIGHT:
		return UPLEFT
	}
	panic("Invalid tile type")
}

func (tile *Tile) IsCompletelyAccessible() bool {
	return tile.Index == 0
}

func (tile *Tile) IsCompletelySolid() bool {
	if tile.IsCompletelyAccessible() || tile.IsDiagonal() {
		return false
	}
	return true
}

func (tile *Tile) IsDiagonal() bool {
	return tile.Index >= FIRST_DIAGONAL_TILE_TYPE
}

func (tile *Tile) GetType() TileType {
	if tile.Index == 0 {
		return COMPLETELY_ACCESSIBLE
	}
	if !tile.IsDiagonal() {
		return COMPLETELY_SOLID
	}

	var FlagLookupTable = []TileType{
		SOLID_AT_UPPER_LEFT, SOLID_AT_UPPER_RIGHT, SOLID_AT_LOWER_LEFT, SOLID_AT_LOWER_RIGHT,
		SOLID_AT_LOWER_RIGHT, SOLID_AT_UPPER_RIGHT, SOLID_AT_LOWER_LEFT, SOLID_AT_UPPER_LEFT,
	}
	return FlagLookupTable[tile.Flags&0x07]
}

// GetRightVector analyses the flags (rotation and flipping), and returns a vector that points upwards, depending on the actual flags. (If rotated 90deg CW, (0,1) is returned)
func (tile *Tile) GetUpVector() (int, int) {
	switch tile.Flags & 0x07 {
	case 0:
		return 0, -1 // up
	case 1:
		return -1, 0 // left
	case 2:
		return 1, 0 // right
	case 3:
		return 0, 1 // down

	case 4:
		return 0, 1 // down
	case 5:
		return 1, 0 // right
	case 6:
		return -1, 0 // left
	case 7:
		return 0, -1 // up
	}
	panic("Invalid state")
}

func (tile *Tile) GetRightVector() (int, int) {
	x, y := tile.GetUpVector()
	return -y, x
}

func PopCount(b uint8) uint8 {
	b = b - ((b >> 1) & 0x55)
	b = (b & 0x33) + ((b >> 2) & 0x33)
	return (((b + (b >> 4)) & 0x0F) * 0x01)
}

func (tile *Tile) IsMirrored() bool {
	return PopCount(tile.Flags&0x07)%2 == 1
}

// Returns true if this tile has a staight border on the top
func (tile *Tile) HasBorderTowards(side Orientation) bool {
	switch tile.GetType() {
	case COMPLETELY_ACCESSIBLE:
		return false
	case COMPLETELY_SOLID:
		return !IsOrientationDiagonal(side)
	case SOLID_AT_UPPER_LEFT:
		return side == LEFT || side == UP || side == DOWNRIGHT
	case SOLID_AT_UPPER_RIGHT:
		return side == RIGHT || side == UP || side == DOWNLEFT
	case SOLID_AT_LOWER_LEFT:
		return side == LEFT || side == DOWN || side == UPRIGHT
	case SOLID_AT_LOWER_RIGHT:
		return side == RIGHT || side == DOWN || side == UPLEFT
	}
	panic("Invalid tile type")
}

func (tilemap *TileMap) GetLayer(layername string) (int, error) {
	layerIdx := -1
	for idx, layer := range tilemap.Layers {
		if layer.Name != layername {
			continue
		}
		if layerIdx == -1 {
			layerIdx = idx
		} else {
			return -1, fmt.Errorf("Multiple layers with name '%v' found", layername)
		}
	}
	if layerIdx == -1 {
		return -1, fmt.Errorf("No layer with name 'layername' found")
	}
	return layerIdx, nil
}

func (tilemap *TileMap) String() string {
	var str = fmt.Sprintf(
		"Version:           %v\n"+
			"Size:              %vx%v\n"+
			"Layer count:       %v\n"+
			"Orientation:       %v\n"+
			"Renderorder:       %v\n"+
			"Tile size:         %vx%v\n",
		tilemap.Version,
		tilemap.Width, tilemap.Height,
		len(tilemap.Layers),
		tilemap.Orientation,
		tilemap.Renderorder,
		tilemap.Tilewidth, tilemap.Tileheight)

	str += "Tilesets:"
	for i, tileset := range tilemap.Tilesets {
		str += fmt.Sprintf("\n\tTileset %d: '%s', firstgid=%d, count=%d", i, tileset.Name, tileset.FirstGid, tileset.TileCount)
	}

	str += "\nLayers:"
	for i, layer := range tilemap.Layers {
		str += fmt.Sprintf("\n\tLayer %d:  '%s'", i, layer.Name)
	}
	return str
}

func (layer *TileMapLayer) String() string {
	return fmt.Sprintf(
		"Layer name:    %v\n",
		layer.Name)
}

func LoadTilesFile(filepath string) (tilemap TileMap, err error) {
	sourceData, err := ioutil.ReadFile(filepath)
	if err != nil {
		return tilemap, fmt.Errorf("Failed to read source file '%v': %v", filepath, err)
	}

	err = xml.Unmarshal(sourceData, &tilemap)
	if err != nil {
		return tilemap, err
	}

	// Validate tilesets and assign types:
	for idx, tileset := range tilemap.Tilesets {
		switch strings.ToLower(tileset.Name) {
		case "environment":
			tilemap.Tilesets[idx].Type = ENVIRONMENT_TILESET
		case "decoration":
			tilemap.Tilesets[idx].Type = DECORATION_TILESET
		case "spawn":
			tilemap.Tilesets[idx].Type = SPAWN_TILESET
		default:
			return tilemap, fmt.Errorf("Failed to read source file '%v': Invalid tilesets detected. The tilset name '%v' is not allowed and must be 'environment', 'decoration' or 'spawn'.", filepath, tileset.Name)
		}
	}

	expectedTileCount := tilemap.Width * tilemap.Height
	for idx := range tilemap.Layers {
		if err := tilemap.Layers[idx].extractTiles(expectedTileCount, tilemap.Tilesets); err != nil {
			return tilemap, err
		}
	}
	return tilemap, err
}

// extractTiles convert's the layers raw data into correct tile data.
func (layer *TileMapLayer) extractTiles(expectedTileCount int, Tilesets []TileSet) error {
	tiles := strings.FieldsFunc(layer.RawData, func(r rune) bool { // remove separators
		return r == ',' || r == '\n' || r == '\r'
	})

	if len(tiles) != expectedTileCount {
		return fmt.Errorf("Unexpected layer data. Tile count doesn't match map size")
	}

	layer.Tiles = make([]Tile, expectedTileCount)

	for i := 0; i < len(tiles); i++ {
		value, err := strconv.Atoi(tiles[i])
		if err != nil {
			return fmt.Errorf("Unexpected layer data. Failed to parse tile number: '%v'", tiles[i])
		}

		tileID := uint32(value)

		var flags uint8 = 0
		if tileID&FlippedHorizontallyTiledFlag != 0 {
			flags |= 0x01
		}
		if tileID&FlippedVerticallyTiledFlag != 0 {
			flags |= 0x02
		}
		if tileID&FlippedDiagonallyTiledFlag != 0 {
			flags |= 0x04
		}
		tileID &^= (FlippedHorizontallyTiledFlag | FlippedVerticallyTiledFlag | FlippedDiagonallyTiledFlag)

		if tileID < 0 || tileID > 0xFFFFFF {
			return fmt.Errorf("Unexpected layer data. Tile number is invalid (additional flag?)")
		}

		// Check which tileset the tile belongs to
		var tileSet *TileSet

		if tileID > 0 {
			for i := 0; i < len(Tilesets) && tileID >= Tilesets[i].FirstGid; i++ {
				tileSet = &Tilesets[i]
			}

			// Check whether the gid is really inside our tilesets
			if tileID >= tileSet.FirstGid+tileSet.TileCount {
				return fmt.Errorf("Unexpected tileID %d. tileID does not belong to any tileset. Last valid id=%d", tileID, tileSet.FirstGid+tileSet.TileCount-1)
			}
		}

		layer.Tiles[i] = Tile{
			Index:   tileID,
			Flags:   flags,
			TileSet: tileSet,
		}
	}

	return nil
}

// GetTile returns the tile at a given position
func (layer *TileMapLayer) GetTile(X, Y, MapWidth, MapHeight int) (tile Tile, err error) {
	if X < 0 || X >= MapWidth || Y < 0 || Y >= MapHeight {
		return tile, fmt.Errorf("Invalid coordinates supplied")
	}
	idx := Y*MapWidth + X
	if idx < 0 || idx >= len(layer.Tiles) {
		return tile, fmt.Errorf("Invalid map size supplied")
	}
	return layer.Tiles[idx], nil
}
