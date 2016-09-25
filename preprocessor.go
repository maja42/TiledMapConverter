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
//     The actual direction is stored with the help of type SortedBorderLines
type BorderLine struct {
    StartX int
    StartY int
    Length int
}

type SortedBorderLines struct {
    Left  []BorderLine // pointing left. solid terrain is above.
    Right []BorderLine // pointing right. solid terrain is below.
    Up    []BorderLine // pointing up. solid terrain is on the right.
    Down  []BorderLine // pointing down. solid terrain is on the left.
}

func (borders *SortedBorderLines) String() string {
    var str = fmt.Sprintf("Number of borders (left, right, up, down): %d, %d, %d, %d",
        len(borders.Left), len(borders.Right), len(borders.Up), len(borders.Down))

    for i, b := range borders.Left {
        str += fmt.Sprintf("\tLeft  %4d: %3d x%3d, length %3d\n", i, b.StartX, b.StartY, b.Length)
    }
    for i, b := range borders.Right {
        str += fmt.Sprintf("\tRight %4d: %3d x%3d, length %3d\n", i, b.StartX, b.StartY, b.Length)
    }
    for i, b := range borders.Up {
        str += fmt.Sprintf("\tUp    %4d: %3d x%3d, length %3d\n", i, b.StartX, b.StartY, b.Length)
    }
    for i, b := range borders.Down {
        str += fmt.Sprintf("\tDown  %4d: %3d x%3d, length %3d\n", i, b.StartX, b.StartY, b.Length)
    }
    return str
}

//
func ComputeBorder(tilemap TileMap) (borders SortedBorderLines, err error) {
    environmentLayerIdx, err := tilemap.GetLayer("environment")
    if err != nil {
        return borders, err
    }

    borders, err = ComputeBorderOfLayer(tilemap.Width, tilemap.Height, tilemap.Layers[environmentLayerIdx])
    return borders, err
}

func ComputeBorderOfLayer(width, height int, layer TileMapLayer) (SortedBorderLines, error) {
    var err error
    var borders = SortedBorderLines{
        Left:  make([]BorderLine, 0, 64),
        Right: make([]BorderLine, 0, 64),
        Up:    make([]BorderLine, 0, 64),
        Down:  make([]BorderLine, 0, 64),
    }

    // Find horizontal borders:
    for y := 1; y < height; y++ {
        var upwardsBorderStart = -1
        var downwardsBorderStart = -1

        for x := 1; x < width; x++ {
            var above Tile
            var mine Tile

            if above, err = layer.GetTile(x, y-1, width, height); err != nil {
                return borders, fmt.Errorf("Failed to compute horizontal border (%dx%d-1): %v", x, y, err)
            }
            if mine, err = layer.GetTile(x, y, width, height); err != nil {
                return borders, fmt.Errorf("Failed to compute horizontal border (%dx%d): %v", x, y, err)
            }

            // Border facing upwards
            if above.Index == 0 && mine.Index != 0 && x != width-1 {
                if upwardsBorderStart == -1 {
                    upwardsBorderStart = x // the border just started
                }
            } else {
                if upwardsBorderStart != -1 { // the border just ended
                    upwardsBorderEnd := x
                    borders.Right = append(borders.Right, BorderLine{ // below = solid
                        StartX: upwardsBorderStart,
                        StartY: y,
                        Length: upwardsBorderEnd - upwardsBorderStart,
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
                    borders.Left = append(borders.Left, BorderLine{ // above = solid
                        StartX: downwardsBorderEnd, // the border goes from right to left (solid region must be on its right side)
                        StartY: y,
                        Length: downwardsBorderEnd - downwardsBorderStart,
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
                return borders, fmt.Errorf("Failed to compute vertical border (%d-1x%d): %v", x, y, err)
            }
            if mine, err = layer.GetTile(x, y, width, height); err != nil {
                return borders, fmt.Errorf("Failed to compute vertical border (%dx%d): %v", x, y, err)
            }

            // Border facing to the left
            if left.Index == 0 && mine.Index != 0 && y != height-1 {
                if leftBorderStart == -1 {
                    leftBorderStart = y // the border just started
                }
            } else {
                if leftBorderStart != -1 { // the border just ended
                    leftBorderEnd := y
                    borders.Up = append(borders.Up, BorderLine{ // right = solid
                        StartX: x,
                        StartY: leftBorderStart,
                        Length: leftBorderEnd - leftBorderStart,
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
                    borders.Down = append(borders.Down, BorderLine{ // left = solid
                        StartX: x,
                        StartY: rightBorderEnd, // the border goes from right to left (solid region must be on its right side)
                        Length: rightBorderEnd - rightBorderStart,
                    })
                    rightBorderStart = -1
                }
            }
        }
    }

    // Possible optimisation: if the map contains unreachable positions, it's borders can be dropped

    // Validate and reduce:
    if len(borders.Left) == 0 || len(borders.Right) == 0 || len(borders.Up) == 0 || len(borders.Down) == 0 {
        return borders, fmt.Errorf("Invalid map: Failed to compute border. A closed map contains at least one border in each direction. "+
            "Found (left, right, up, down): %d, %d, %d, %d ",
            len(borders.Left), len(borders.Right), len(borders.Up), len(borders.Down))
    }

    return borders, nil
}
