package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
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
	if err := Run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	log.Info("Success")
}

// Run executes the application and returns an error message if something went wrong
func Run() error {
	SetupLogger(logging.INFO)

	if len(os.Args) != 2 {
		return fmt.Errorf("Usage: %s <inputfile.tmx>", os.Args[0])
	}

	var sourceFile = os.Args[1]
	var targetFile = GetTargetFilePath(sourceFile)

	tilemap, err := LoadTilesFile(sourceFile)
	if err != nil {
		return fmt.Errorf("Failed to load source file: %v", err)
	}

	log.Info("Input data:\n" + tilemap.String())
	log.Infof("---------------------------------------")

	if err := ValidateTileMap(tilemap); err != nil {
		return err
	}

	resources, players, err := ExtractSpawnInfo(tilemap)
	if err != nil {
		return err
	}

	borders, err := ComputeBorder(tilemap)
	if err != nil {
		return err
	}

	log.Infof("Number of resource points: %d", len(resources))
	for i, r := range resources {
		log.Infof("\t%2d: %3d x%3d", i, r.SpawnX, r.SpawnY)
	}

	log.Infof("Number of players: %d", len(players))
	for i, p := range players {
		log.Infof("\tPlayer %d: %3d x%3d, %d units", i, p.SpawnX, p.SpawnY, len(p.Units))
	}

	log.Infof("Number of borders (left, right, up, down): %d, %d, %d, %d",
		len(borders.Left), len(borders.Right), len(borders.Up), len(borders.Down))
	log.Debug(borders.String())

	log.Infof("Writing to '%s'", targetFile)
	err = os.Remove(targetFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Failed to remove existing file '%v'", targetFile)
	}

	file, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("Failed to create output file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	err = Encode(writer, binary.LittleEndian, tilemap, resources, players, borders)
	if err != nil {
		os.Remove(targetFile)
		return fmt.Errorf("Failed to write output file: %v", err)
	}
	writer.Flush()
	return nil
}
