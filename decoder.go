package main

import (
    "encoding/xml"
    "fmt"
    "io/ioutil"
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

type TileMapLayer struct {
    Name string `xml:"name,attr"`
    Data string `xml:"data"`
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
        str += fmt.Sprintf("\nLayer %d:\n%s", i, layer.String())
    }
    return str
}

func (layer *TileMapLayer) String() string {
    return fmt.Sprintf(
        "\tLayer name:    %v\n",
        layer.Name)
}

func LoadTilesFile(filepath string) (tilemap TileMap, err error) {
    sourceData, err := ioutil.ReadFile(filepath)
    if err != nil {
        return tilemap, fmt.Errorf("Failed to read source file '%v': %v", filepath, err)
    }

    err = xml.Unmarshal(sourceData, &tilemap)
    return tilemap, err
}
