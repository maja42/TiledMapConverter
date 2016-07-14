package main

import (
    "bufio"
    "encoding/binary"
    "fmt"
    "strconv"
    "strings"
)

func encode(writer *bufio.Writer, order binary.ByteOrder, tilemap TileMap) error {
    if err := binary.Write(writer, order, uint32(tilemap.Width)); err != nil {
        return err
    }
    if err := binary.Write(writer, order, uint32(tilemap.Height)); err != nil {
        return err
    }

    expectedTileCount := tilemap.Width * tilemap.Height

    for _, layer := range tilemap.Layers {
        // tiles := strings.Split(layer.Data, ",\n\r")
        tiles := strings.FieldsFunc(layer.Data, func(r rune) bool {
            return r == ',' || r == '\n' || r == '\r'
        })

        if len(tiles) != expectedTileCount {
            return fmt.Errorf("Unexpected layer data. Tile count doesn't match map size")
        }
        for i := 0; i < len(tiles); i++ {
            var tileNum int
            tileNum, err := strconv.Atoi(tiles[i])
            if err != nil {
                return fmt.Errorf("Unexpected layer data. Failed to parse tile number: '%v'", tiles[i])
            }
            if tileNum < 0 || tileNum > 255 {
                return fmt.Errorf("Unexpected layer data. Tile number can't be encoded (not within range [0,256]): %d", tileNum)
            }
            writer.WriteByte(byte(uint8(tileNum)))
        }
    }
    return nil
}
