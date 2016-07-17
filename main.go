package main

import (
	"bufio"
	"encoding/binary"
	"github.com/op/go-logging"
	"os"
	"path/filepath"
)

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
		return
	}

	var sourceFile = os.Args[1]
	var targetFile = GetTargetFilePath(sourceFile)

	tilemap, err := LoadTilesFile(sourceFile)
	if err != nil {
		log.Errorf("Failed to load source file: %v", err)
		return
	}

	log.Info(tilemap.String())

	if err := ValidateTileMap(tilemap); err != nil {
		log.Error(err)
		return
	}

	players, err := ExtractSpawnInfo(tilemap)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("Number of players: %d", len(players))
	for i, p := range players {
		log.Infof("\tPlayer %d: %3d x%3d, %d units", i, p.SpawnX, p.SpawnY, len(p.Units))
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
	err = Encode(writer, binary.LittleEndian, tilemap, players)
	if err != nil {
		log.Errorf("Failed to write output file: %v", err)
		os.Remove(targetFile)
		return
	}
	writer.Flush()

	log.Info("Success")
}
