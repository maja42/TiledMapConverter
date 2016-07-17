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

    Layers []TileMapLayer `xml:"layer"`
}

const (
    FlippedHorizontallyTiledFlag uint32 = 0x80000000
    FlippedVerticallyTiledFlag   uint32 = 0x40000000
    FlippedDiagonallyTiledFlag   uint32 = 0x20000000
)

type TileMapLayer struct {
    Name    string `xml:"name,attr"`
    RawData string `xml:"data"`
    Tiles   []Tile `xml:"-"`
}

type Tile struct {
    Index uint32
    Flags uint8
}

func (tileMap *TileMap) String() string {
    var str = fmt.Sprintf(
        "Version:           %v\n"+
            "Size:              %vx%v\n"+
            "Layer count:       %v\n"+
            "Orientation:       %v\n"+
            "Renderorder:       %v\n"+
            "Tile size:         %vx%v",
        tileMap.Version,
        tileMap.Width, tileMap.Height,
        len(tileMap.Layers),
        tileMap.Orientation,
        tileMap.Renderorder,
        tileMap.Tilewidth, tileMap.Tileheight)

    for i, layer := range tileMap.Layers {
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

    expectedTileCount := tilemap.Width * tilemap.Height
    for idx := range tilemap.Layers {
        if err := tilemap.Layers[idx].extractTiles(expectedTileCount); err != nil {
            return tilemap, err
        }
    }
    return tilemap, err
}

func (layer *TileMapLayer) extractTiles(expectedTileCount int) error {
    tiles := strings.FieldsFunc(layer.RawData, func(r rune) bool {
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

        layer.Tiles[i] = Tile{
            Index: tileID,
            Flags: flags,
        }
    }

    return nil
}
