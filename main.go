package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/op/go-logging"
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
	SetupLogger(logging.DEBUG)

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

	if err := ValidateTileMap(&tilemap); err != nil {
		return err
	}

	resources, waterdropSources, players, err := ExtractSpawnInfo(&tilemap)
	if err != nil {
		return err
	}

	borders, err := ComputeBorder(&tilemap)
	if err != nil {
		return err
	}

	log.Infof("Number of resource points: %d", len(resources))
	// for i, r := range resources {
	// 	log.Debugf("\t%2d: %3d x%3d", i, r.SpawnX, r.SpawnY)
	// }

	log.Infof("Number of water drop sources: %d", len(waterdropSources))
	// for i, s := range waterdropSources {
	// 	log.Debugf("\t%2d: %3d x%3d", i, s.SpawnX, s.SpawnY)
	// }

	log.Infof("Number of players: %d", len(players))
	for i, p := range players {
		log.Infof("\tPlayer %d: %d buildings, %d units", i, len(p.Buildings), len(p.Units))
	}

	objectCount := 0
	if tilemap.ForegroundObjectLayer != nil {
		objectCount = len(tilemap.ForegroundObjectLayer.Objects)
	}
	log.Infof("Number of objects in foreground object layer: %d", objectCount)
	objectCount = 0
	if tilemap.BackgroundObjectLayer != nil {
		objectCount = len(tilemap.BackgroundObjectLayer.Objects)
	}
	log.Infof("Number of objects in background object layer: %d", objectCount)

	log.Infof("Number of borders (left, right, up, down): %d, %d, %d, %d",
		len(borders.Left), len(borders.Right), len(borders.Up), len(borders.Down))
	log.Infof("Number of borders (up-left, up-right, down-left, down-right): %d, %d, %d, %d",
		len(borders.UpLeft), len(borders.UpRight), len(borders.DownLeft), len(borders.DownRight))
	//log.Debug(borders.String())

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
	err = Encode(writer, binary.LittleEndian, &tilemap, resources, waterdropSources, players, borders)
	if err != nil {
		os.Remove(targetFile)
		return fmt.Errorf("Failed to write output file: %v", err)
	}
	writer.Flush()
	return nil
}
