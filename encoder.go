package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
)

// Encode encodes and writes the given tilemap into the writer (=output file)
func Encode(writer *bufio.Writer, order binary.ByteOrder, tilemap *TileMap, resourcePoints []ResourcePoint, waterdropSources []WaterdropSource, players []Player, borders SortedBorderLines) error {
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

	if err := encodeObjectLayer(writer, order, tilemap.BackgroundObjectLayer); err != nil {
		return fmt.Errorf("Failed to encode BackgroundObjectLayer: %v", err)
	}
	if err := encodeObjectLayer(writer, order, tilemap.ForegroundObjectLayer); err != nil {
		return fmt.Errorf("Failed to encode ForegroundObjectLayer: %v", err)
	}

	writer.WriteByte(byte(0x99)) // magic byte

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

	if len(waterdropSources) < 0 || len(waterdropSources) > 0xFF {
		return fmt.Errorf("Number of water drop sources can't be encoded (not within range [0,256]): %d", len(waterdropSources))
	}
	writer.WriteByte(byte(uint8(len(waterdropSources)))) // number of water drop sources
	for _, source := range waterdropSources {
		if err := encodeWaterdropSource(writer, order, &source); err != nil {
			return err
		}
	}
	writer.WriteByte(byte(0xFF)) // magic byte

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
	tilesetType := probeLayer(layer)
	writer.WriteByte(byte(tilesetType))

	for i, tile := range layer.Tiles {
		tileID := tile.Index

		if tileID > 0 && tile.TileSet.Type != tilesetType {
			return fmt.Errorf("The tile (%d, layer=%q) can't be encoded. All tiles within a layer must come from the same tileset.", i, layer.Name)
		}

		if tileID < 0 || tileID > 0xFF {
			return fmt.Errorf("Tile index can't be encoded (not within range [0,256]): %d", tileID)
		}

		writer.WriteByte(byte(tile.Flags))
		writer.WriteByte(byte(uint8(tileID)))

	}
	return nil
}

// probeLayer goes through all tiles and returns the tileset-type of the first occupied tile it finds
func probeLayer(layer *TileMapLayer) TileSetType {
	for _, tile := range layer.Tiles {
		if tile.Index > 0 {
			return tile.TileSet.Type
		}
	}
	log.Warningf("The layer %q is completely empty and should be removed", layer.Name)
	return DECORATION1_TILESET
}

func encodeObjectLayer(writer *bufio.Writer, order binary.ByteOrder, layer *TileMapObjectLayer) error {
	var objectCount int = 0
	if layer != nil {
		objectCount = len(layer.Objects)
	}
	if objectCount < 0 || objectCount > 0xFFFF {
		return fmt.Errorf("Number of objects can't be encoded (16bit): %d", objectCount)
	}

	if err := binary.Write(writer, order, int16(objectCount)); err != nil {
		return err
	}

	if layer == nil {
		return nil
	}

	for i, object := range layer.Objects {
		if object.TileSet == nil {
			return fmt.Errorf("The object (%d, layer=%q) can't be encoded. No valid tileset.", i, layer.Name)
		} else if object.TileSet.Type != DECORATION1_TILESET {
			return fmt.Errorf("Unsupported object tileset (%d, layer=%q). Only the decoration tileset 1 can be used for object layers", i, layer.Name)
		}

		tileID := object.Index

		if tileID < 0 || tileID > 0xFF {
			return fmt.Errorf("Tile index of object can't be encoded (not within range [0,256]): %d", tileID)
		}

		writer.WriteByte(byte(uint8(tileID)))

		// Tiled uses the bottom-left corner for the position. We store the object's center ==> convert!
		localCenterX := object.Width / 2
		localCenterY := object.Height / 2
		cosRot := float32(math.Cos(float64(-object.Rotation) / 180 * math.Pi))
		sinRot := float32(math.Sin(float64(-object.Rotation) / 180 * math.Pi))

		rotatedCenterX := localCenterX*cosRot - localCenterY*sinRot
		rotatedCenterY := localCenterX*sinRot + localCenterY*cosRot

		centerX := object.X + rotatedCenterX
		centerY := object.Y - rotatedCenterY // objects have an inverted coordinate system (up = positive)

		Hor := (object.Flags & 0x01) != 0
		Ver := (object.Flags & 0x02) != 0
		Diag := (object.Flags & 0x04) != 0

		if Hor {
			object.Width = -object.Width
		}
		if Ver {
			object.Height = -object.Height
		}
		if Diag {
			return fmt.Errorf("Unable to encode object (%d, layer=%q) - Unexpected flag. Tiled should not set the diagonal-flipped flag, as such flips can always be expressed with X/Y-flips and rotations", i, layer.Name)
		}

		if err := writeFloat(writer, order, centerX/float32(object.TileSet.TileWidth)); err != nil {
			return fmt.Errorf("Unable to encode object (%d, layer=%q) - Failed to write x-coordinate: %v", i, layer.Name, err)
		}
		if err := writeFloat(writer, order, centerY/float32(object.TileSet.TileWidth)); err != nil { // invert y axis
			return fmt.Errorf("Unable to encode object (%d, layer=%q) - Failed to write y-coordinate: %v", i, layer.Name, err)
		}
		if err := writeFloat(writer, order, object.Width/float32(object.TileSet.TileHeight)); err != nil {
			return fmt.Errorf("Unable to encode object (%d, layer=%q) - Failed to write width: %v", i, layer.Name, err)
		}
		if err := writeFloat(writer, order, object.Height/float32(object.TileSet.TileHeight)); err != nil {
			return fmt.Errorf("Unable to encode object (%d, layer=%q) - Failed to write height: %v", i, layer.Name, err)
		}
		if err := writeFloat(writer, order, object.Rotation); err != nil {
			return fmt.Errorf("Unable to encode object (%d, layer=%q) - Failed to write rotation: %v", i, layer.Name, err)
		}
	}
	return nil
}

func writeFloat(writer *bufio.Writer, order binary.ByteOrder, value float32) error {
	var intVal int = int(value * 1000) // All floats are multiplied by 1000. The loader has to divide by 1000 to get the original float value.
	return binary.Write(writer, order, int32(intVal))
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

func encodeWaterdropSource(writer *bufio.Writer, order binary.ByteOrder, source *WaterdropSource) error {
	if err := binary.Write(writer, order, int16(source.SpawnX)); err != nil {
		return err
	}
	if err := binary.Write(writer, order, int16(source.SpawnY)); err != nil {
		return err
	}
	writer.WriteByte(byte(source.WaterdropFlags))
	return nil
}

func encodePlayer(writer *bufio.Writer, order binary.ByteOrder, player *Player) error {
	if err := encodeBuildings(writer, order, player); err != nil {
		return err
	}
	if err := encodeUnits(writer, order, player); err != nil {
		return err
	}
	return nil
}

func encodeBuildings(writer *bufio.Writer, order binary.ByteOrder, player *Player) error {
	buildingCount := len(player.Buildings)
	if buildingCount < 0 || buildingCount > 0xFF {
		return fmt.Errorf("Player buildings can't be encoded (building count not within range [0,256]): %d", buildingCount)
	}

	writer.WriteByte(byte(buildingCount)) // Building count

	for _, building := range player.Buildings {
		if building.Type < 0 || building.Type > 0xFF {
			return fmt.Errorf("Building can't be encoded (building type not within range [0,256]): %d", building.Type)
		}

		writer.WriteByte(byte(building.Type))

		if err := binary.Write(writer, order, int16(building.SpawnX)); err != nil {
			return err
		}
		if err := binary.Write(writer, order, int16(building.SpawnY)); err != nil {
			return err
		}

		writer.WriteByte(byte(building.Flags))
	}
	return nil
}

func encodeUnits(writer *bufio.Writer, order binary.ByteOrder, player *Player) error {
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

	if err := binary.Write(writer, order, int16(len(borders.UpLeft))); err != nil {
		return err
	}
	if err := binary.Write(writer, order, int16(len(borders.UpRight))); err != nil {
		return err
	}
	if err := binary.Write(writer, order, int16(len(borders.DownLeft))); err != nil {
		return err
	}
	if err := binary.Write(writer, order, int16(len(borders.DownRight))); err != nil {
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

	for _, line := range borders.UpLeft {
		if err := encodeBorderLine(writer, order, line); err != nil {
			return err
		}
	}
	for _, line := range borders.UpRight {
		if err := encodeBorderLine(writer, order, line); err != nil {
			return err
		}
	}
	for _, line := range borders.DownLeft {
		if err := encodeBorderLine(writer, order, line); err != nil {
			return err
		}
	}
	for _, line := range borders.DownRight {
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
