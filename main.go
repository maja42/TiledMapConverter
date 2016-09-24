package main

import (
	"bufio"
	"encoding/binary"
	"github.com/op/go-logging"
	"os"
	"path/filepath"
)

// GetTargetFilePath returns the file path for the new, converted file that has the same name/path as the input file
func GetTargetFilePath(sourceFile string) string {
	path, filename := filepath.Split(sourceFile)
	ext := filepath.Ext(filename)
	filename = filename[:len(filename)-len(ext)]
	return path + filename + ".tilemap"
}

func main() {
	SetupLogger(logging.INFO)

	if len(os.Args) != 2 {
		log.Errorf("Usage: %s <inputfile.tmx>", os.Args[0])
		os.Exit(1)
		return
	}

	var sourceFile = os.Args[1]
	var targetFile = GetTargetFilePath(sourceFile)

	tilemap, err := LoadTilesFile(sourceFile)
	if err != nil {
		log.Errorf("Failed to load source file: %v", err)
		return
	}

	log.Info("Input data:\n" + tilemap.String())
	log.Infof("---------------------------------------")

	if err := ValidateTileMap(tilemap); err != nil {
		log.Error(err)
		return
	}

	resources, players, err := ExtractSpawnInfo(tilemap)
	if err != nil {
		log.Error(err)
		return
	}

	borders, err := ComputeBorder(tilemap)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("Number of resource points: %d", len(resources))
	for i, r := range resources {
		log.Infof("\t%2d: %3d x%3d", i, r.SpawnX, r.SpawnY)
	}

	log.Infof("Number of players: %d", len(players))
	for i, p := range players {
		log.Infof("\tPlayer %d: %3d x%3d, %d units", i, p.SpawnX, p.SpawnY, len(p.Units))
	}

	log.Infof("Number of borders: %d", len(borders))
	for i, b := range borders {
		log.Debugf("\t%4d: %3d x%3d --> %3d x%3d", i, b.StartX, b.StartY, b.EndX, b.EndY)
	}

	log.Infof("Writing to '%s'", targetFile)
	err = os.Remove(targetFile)
	if err != nil && !os.IsNotExist(err) {
		log.Errorf("Failed to remove existing file '%v'", targetFile)
		return
	}

	file, err := os.Create(targetFile)
	if err != nil {
		log.Errorf("Failed to create output file: %v", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	err = Encode(writer, binary.LittleEndian, tilemap, resources, players, borders)
	if err != nil {
		log.Errorf("Failed to write output file: %v", err)
		os.Remove(targetFile)
		return
	}
	writer.Flush()

	log.Info("Success")
}
