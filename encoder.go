package main

import (
    "bufio"
    "encoding/binary"
    "fmt"
)

// Encode encodes and writes the given tilemap into the writer (=output file)
func Encode(writer *bufio.Writer, order binary.ByteOrder, tilemap TileMap, resourcePoints []ResourcePoint, players []Player, borders SortedBorderLines) error {
    writer.WriteByte(byte(0xA5)) // magic byte
    writer.WriteByte(byte(0x02)) // magic byte used for versioning

    if err := binary.Write(writer, order, int16(tilemap.Width)); err != nil {
        return err
    }
    if err := binary.Write(writer, order, int16(tilemap.Height)); err != nil {
        return err
    }
    writer.WriteByte(byte(uint8(len(tilemap.Layers))))

    environmentLayerIdx, err := tilemap.GetLayer("environment")
    if err != nil {
        return err
    }
    environmentLayerIdx = len(tilemap.Layers) - 1 - environmentLayerIdx // The layers will be stored in reversed order
    writer.WriteByte(byte(environmentLayerIdx))

    for i := len(tilemap.Layers) - 1; i >= 0; i-- {
        layer := tilemap.Layers[i]
        if err := encodeLayer(writer, order, &layer); err != nil {
            return err
        }
    }
    writer.WriteByte(byte(0xAA)) // magic byte

    if len(resourcePoints) < 0 || len(resourcePoints) > 0xFF {
        return fmt.Errorf("Number of resource points can't be encoded (not within range [0,256]): %d", len(resourcePoints))
    }
    writer.WriteByte(byte(uint8(len(resourcePoints)))) // number of resource points
    for _, resource := range resourcePoints {
        if err := encodeResourcePoint(writer, order, &resource); err != nil {
            return err
        }
    }
    writer.WriteByte(byte(0x5A)) // magic byte

    writer.WriteByte(byte(uint8(len(players)))) // number of players
    for _, player := range players {
        if err := encodePlayer(writer, order, &player); err != nil {
            return err
        }
    }

    writer.WriteByte(byte(0xA5)) // magic byte
    if err := encodeBorders(writer, order, borders); err != nil {
        return err
    }

    writer.WriteByte(byte(0x55)) // magic byte
    return nil
}

func encodeLayer(writer *bufio.Writer, order binary.ByteOrder, layer *TileMapLayer) error {
    for _, tile := range layer.Tiles {
        if tile.Index < 0 || tile.Index > 0xFF {
            return fmt.Errorf("Tile index can't be encoded (not within range [0,256]): %d", tile.Index)
        }
        writer.WriteByte(byte(tile.Flags))
        writer.WriteByte(byte(uint8(tile.Index)))
    }
    return nil
}

func encodeResourcePoint(writer *bufio.Writer, order binary.ByteOrder, resource *ResourcePoint) error {
    if err := binary.Write(writer, order, int16(resource.SpawnX)); err != nil {
        return err
    }
    if err := binary.Write(writer, order, int16(resource.SpawnY)); err != nil {
        return err
    }
    writer.WriteByte(byte(resource.ResourcePointFlags))
    return nil
}

func encodePlayer(writer *bufio.Writer, order binary.ByteOrder, player *Player) error {
    if err := binary.Write(writer, order, int16(player.SpawnX)); err != nil {
        return err
    }
    if err := binary.Write(writer, order, int16(player.SpawnY)); err != nil {
        return err
    }
    writer.WriteByte(byte(player.BaseBuildingFlags))

    unitCount := len(player.Units)
    if unitCount < 0 || unitCount > 0xFF {
        return fmt.Errorf("Player units can't be encoded (unit count not within range [0,256]): %d", unitCount)
    }

    writer.WriteByte(byte(unitCount)) // Unit count

    for _, unit := range player.Units {
        if unit.Type < 0 || unit.Type > 0xFF {
            return fmt.Errorf("Unit can't be encoded (unit type not within range [0,256]): %d", unit.Type)
        }

        writer.WriteByte(byte(unit.Type))
        if err := binary.Write(writer, order, int16(unit.SpawnX)); err != nil {
            return err
        }
        if err := binary.Write(writer, order, int16(unit.SpawnY)); err != nil {
            return err
        }
    }
    return nil
}

func encodeBorders(writer *bufio.Writer, order binary.ByteOrder, borders SortedBorderLines) error {
    if err := binary.Write(writer, order, int16(len(borders.Left))); err != nil {
        return err
    }
    if err := binary.Write(writer, order, int16(len(borders.Right))); err != nil {
        return err
    }
    if err := binary.Write(writer, order, int16(len(borders.Up))); err != nil {
        return err
    }
    if err := binary.Write(writer, order, int16(len(borders.Down))); err != nil {
        return err
    }

    for _, line := range borders.Left {
        if err := encodeBorderLine(writer, order, line); err != nil {
            return err
        }
    }
    for _, line := range borders.Right {
        if err := encodeBorderLine(writer, order, line); err != nil {
            return err
        }
    }
    for _, line := range borders.Up {
        if err := encodeBorderLine(writer, order, line); err != nil {
            return err
        }
    }
    for _, line := range borders.Down {
        if err := encodeBorderLine(writer, order, line); err != nil {
            return err
        }
    }
    return nil
}

func encodeBorderLine(writer *bufio.Writer, order binary.ByteOrder, borderLine BorderLine) error {
    if err := binary.Write(writer, order, int16(borderLine.StartX)); err != nil {
        return err
    }
    if err := binary.Write(writer, order, int16(borderLine.StartY)); err != nil {
        return err
    }
    if err := binary.Write(writer, order, int16(borderLine.Length)); err != nil {
        return err
    }
    return nil
}
