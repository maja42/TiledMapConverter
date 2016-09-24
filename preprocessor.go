package main

import (
    "fmt"
)

// BorderLine represents a border between solid terrain data and air.
// This information is pre-computed and used by the game for easier collision-detection.
//     The left side of the border is always empty, while the right side is solid.
//     A position of 0,0 represents a border in the left-upper corner of the upper-left most tile.
//     Since borders facing to the outside environment (outside the map) are invalid, all positions
//     must be in the range [1, size-1], incl. A map of size 10 can therefore have borders of [1, 9].
type BorderLine struct {
    StartX int
    StartY int
    EndX   int
    EndY   int
}

//
func ComputeBorder(tilemap TileMap) ([]BorderLine, error) {
    environmentLayerIdx, err := tilemap.GetLayer("environment")
    if err != nil {
        return nil, err
    }

    borders, err := ComputeBorderOfLayer(tilemap.Width, tilemap.Height, tilemap.Layers[environmentLayerIdx])
    if err != nil {
        return nil, err
    }
    return borders, nil
}

func ComputeBorderOfLayer(width, height int, layer TileMapLayer) ([]BorderLine, error) {
    var lines = make([]BorderLine, 0, 64)
    var err error

    // Find horizontal borders:
    for y := 1; y < height; y++ {
        var upwardsBorderStart = -1
        var downwardsBorderStart = -1

        for x := 1; x < width; x++ {
            var above Tile
            var mine Tile

            if above, err = layer.GetTile(x, y-1, width, height); err != nil {
                return nil, fmt.Errorf("Failed to compute horizontal border (%dx%d-1): %v", x, y, err)
            }
            if mine, err = layer.GetTile(x, y, width, height); err != nil {
                return nil, fmt.Errorf("Failed to compute horizontal border (%dx%d): %v", x, y, err)
            }

            // Border facing upwards
            if above.Index == 0 && mine.Index != 0 && x != width-1 {
                if upwardsBorderStart == -1 {
                    upwardsBorderStart = x // the border just started
                }
            } else {
                if upwardsBorderStart != -1 { // the border just ended
                    upwardsBorderEnd := x
                    lines = append(lines, BorderLine{
                        StartX: upwardsBorderStart,
                        StartY: y,
                        EndX:   upwardsBorderEnd,
                        EndY:   y,
                    })
                    upwardsBorderStart = -1
                }
            }

            // Border facing downwards
            if above.Index != 0 && mine.Index == 0 && x != width-1 {
                if downwardsBorderStart == -1 {
                    downwardsBorderStart = x // the border just started
                }
            } else {
                if downwardsBorderStart != -1 { // the border just ended
                    downwardsBorderEnd := x
                    lines = append(lines, BorderLine{
                        StartX: downwardsBorderEnd, // the border goes from right to left (solid region must be on its right side)
                        StartY: y,
                        EndX:   downwardsBorderStart,
                        EndY:   y,
                    })
                    downwardsBorderStart = -1
                }
            }
        }
    }

    // Find vertical borders:
    for x := 1; x < width; x++ {
        var leftBorderStart = -1
        var rightBorderStart = -1

        for y := 1; y < height; y++ {
            var left Tile
            var mine Tile

            if left, err = layer.GetTile(x-1, y, width, height); err != nil {
                return nil, fmt.Errorf("Failed to compute vertical border (%d-1x%d): %v", x, y, err)
            }
            if mine, err = layer.GetTile(x, y, width, height); err != nil {
                return nil, fmt.Errorf("Failed to compute vertical border (%dx%d): %v", x, y, err)
            }

            // Border facing to the left
            if left.Index == 0 && mine.Index != 0 && y != height-1 {
                if leftBorderStart == -1 {
                    leftBorderStart = y // the border just started
                }
            } else {
                if leftBorderStart != -1 { // the border just ended
                    leftBorderEnd := y
                    lines = append(lines, BorderLine{
                        StartX: x,
                        StartY: leftBorderStart,
                        EndX:   x,
                        EndY:   leftBorderEnd,
                    })
                    leftBorderStart = -1
                }
            }

            // Border facing to the right
            if left.Index != 0 && mine.Index == 0 && y != height-1 {
                if rightBorderStart == -1 {
                    rightBorderStart = y // the border just started
                }
            } else {
                if rightBorderStart != -1 { // the border just ended
                    rightBorderEnd := y
                    lines = append(lines, BorderLine{
                        StartX: x,
                        StartY: rightBorderEnd, // the border goes from right to left (solid region must be on its right side)
                        EndX:   x,
                        EndY:   rightBorderStart,
                    })
                    rightBorderStart = -1
                }
            }
        }
    }

    // Possible optimisation: if the map contains unreachable positions, it's borders can be dropped

    // Validate and reduce:
    if len(lines) < 4 {
        return nil, fmt.Errorf("Invalid map: Failed to compute border. A closed map contains at least 4 borders, found %d", len(lines))
    }
    return lines, nil
}
