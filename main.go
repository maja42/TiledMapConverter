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

func Run() int {
	SetupLogger(logging.INFO)

	if len(os.Args) != 2 {
		log.Errorf("Usage: %s <inputfile.tmx>", os.Args[0])
		return 1
	}

	var sourceFile = os.Args[1]
	var targetFile = GetTargetFilePath(sourceFile)

	tilemap, err := LoadTilesFile(sourceFile)
	if err != nil {
		log.Errorf("Failed to load source file: %v", err)
		return 1
	}

	log.Info(tilemap.String())

	if err := ValidateTileMap(tilemap); err != nil {
		log.Error(err)
		return 1
	}

	players, err := ExtractSpawnInfo(tilemap)
	if err != nil {
		log.Error(err)
		return 1
	}

	log.Infof("Number of players: %d", len(players))
	for i, p := range players {
		log.Infof("\tPlayer %d: %3d x%3d, %d units", i, p.SpawnX, p.SpawnY, len(p.Units))
	}

	log.Infof("Writing to '%s'", targetFile)
	err = os.Remove(targetFile)
	if err != nil && !os.IsNotExist(err) {
		log.Errorf("Failed to remove existing file '%v'", targetFile)
		return 1
	}

	file, err := os.Create(targetFile)
	if err != nil {
		log.Errorf("Failed to create output file: %v", err)
		return 1
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	err = Encode(writer, binary.LittleEndian, tilemap, players)
	if err != nil {
		log.Errorf("Failed to write output file: %v", err)
		os.Remove(targetFile)
		return 1
	}
	writer.Flush()

	log.Info("Success")
	return 0;
}

func main() {
	os.Exit(Run())
}
