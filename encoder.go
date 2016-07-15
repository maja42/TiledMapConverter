package main

import (
    "bufio"
    "encoding/binary"
    "fmt"
    "strconv"
    "strings"
)

const (
    FLIPPED_HORIZONTALLY_FLAG uint32 = 0x80000000
    FLIPPED_VERTICALLY_FLAG   uint32 = 0x40000000
    FLIPPED_DIAGONALLY_FLAG   uint32 = 0x20000000
)

func encode(writer *bufio.Writer, order binary.ByteOrder, tilemap TileMap) error {
    if err := binary.Write(writer, order, uint32(tilemap.Width)); err != nil {
        return err
    }
    if err := binary.Write(writer, order, uint32(tilemap.Height)); err != nil {
        return err
    }
    writer.WriteByte(byte(uint8(len(tilemap.Layers))))

    expectedTileCount := tilemap.Width * tilemap.Height

    for i := len(tilemap.Layers) - 1; i >= 0; i-- {
        layer := tilemap.Layers[i]
        tiles := strings.FieldsFunc(layer.Data, func(r rune) bool {
            return r == ',' || r == '\n' || r == '\r'
        })

        if len(tiles) != expectedTileCount {
            return fmt.Errorf("Unexpected layer data. Tile count doesn't match map size")
        }
        for i := 0; i < len(tiles); i++ {
            value, err := strconv.Atoi(tiles[i])
            if err != nil {
                return fmt.Errorf("Unexpected layer data. Failed to parse tile number: '%v'", tiles[i])
            }
            var tileNum = uint32(value)

            var flags uint8 = 0
            if tileNum&FLIPPED_HORIZONTALLY_FLAG != 0 {
                flags |= 0x01
            }
            if tileNum&FLIPPED_VERTICALLY_FLAG != 0 {
                flags |= 0x02
            }
            if tileNum&FLIPPED_DIAGONALLY_FLAG != 0 {
                flags |= 0x04
            }
            tileNum &^= (FLIPPED_HORIZONTALLY_FLAG | FLIPPED_VERTICALLY_FLAG | FLIPPED_DIAGONALLY_FLAG)

            if tileNum < 0 || tileNum > 255 {
                return fmt.Errorf("Unexpected layer data. Tile number can't be encoded (not within range [0,256]): %d", tileNum)
            }
            writer.WriteByte(byte(flags))
            writer.WriteByte(byte(uint8(tileNum)))
        }
    }
    return nil
}
