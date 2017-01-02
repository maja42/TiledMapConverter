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

// SortedBorderLines is a collection of multiple border lines, sorted by their direction
type SortedBorderLines struct {
	Left  []BorderLine // pointing left. solid terrain is above.
	Right []BorderLine // pointing right. solid terrain is below.
	Up    []BorderLine // pointing up. solid terrain is on the right.
	Down  []BorderLine // pointing down. solid terrain is on the left.

	UpLeft    []BorderLine // pointing up-left. solid terrain is right-above.
	UpRight   []BorderLine // pointing up-right. solid terrain is right-below.
	DownLeft  []BorderLine // pointing down-left. solid terrain is on the left-above.
	DownRight []BorderLine // pointing down-right. solid terrain is on the left-below.
}

func (borders *SortedBorderLines) String() string {
	var str = fmt.Sprintf("Number of borders (left, right, up, down): %d, %d, %d, %d\n",
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

	str += fmt.Sprintf("Number of borders (up-left, up-right, down-left, down-right): %d, %d, %d, %d\n",
		len(borders.UpLeft), len(borders.UpRight), len(borders.DownLeft), len(borders.DownRight))
	for i, b := range borders.UpLeft {
		str += fmt.Sprintf("\tUpLeft     %4d: %3d x%3d, length %3d\n", i, b.StartX, b.StartY, b.Length)
	}
	for i, b := range borders.UpRight {
		str += fmt.Sprintf("\tUpRight    %4d: %3d x%3d, length %3d\n", i, b.StartX, b.StartY, b.Length)
	}
	for i, b := range borders.DownLeft {
		str += fmt.Sprintf("\tDownLeft   %4d: %3d x%3d, length %3d\n", i, b.StartX, b.StartY, b.Length)
	}
	for i, b := range borders.DownRight {
		str += fmt.Sprintf("\tDownRight  %4d: %3d x%3d, length %3d\n", i, b.StartX, b.StartY, b.Length)
	}
	return str
}

func HasBorderTowards(tile Tile, neighbour Tile, tileSide Orientation) bool {
	if IsOrientationDiagonal(tileSide) {
		panic("Invalid function call. If a tile is diagonal, it always has a border on the diagonal side, independent of its neighbour")
	}

	if !tile.HasBorderTowards(tileSide) {
		return false
	}
	neighbourSide := GetInvertedOrientation(tileSide)
	if neighbour.HasBorderTowards(neighbourSide) { // If 'tile' has a border towards the right, and the right neighbour also has a border on its left side, there is no border.
		return false
	}
	return true
}

func ComputeBorder(tilemap *TileMap) (borders SortedBorderLines, err error) {
	environmentLayerIdx, err := tilemap.GetLayer("environment")
	if err != nil {
		return borders, err
	}

	borders, err = ComputeBorderOfLayer(tilemap.Width, tilemap.Height, &tilemap.Layers[environmentLayerIdx])
	return borders, err
}

func ComputeBorderOfLayer(width, height int, layer *TileMapLayer) (SortedBorderLines, error) {
	var err error
	var borders = SortedBorderLines{
		Left:  make([]BorderLine, 0, 64),
		Right: make([]BorderLine, 0, 64),
		Up:    make([]BorderLine, 0, 64),
		Down:  make([]BorderLine, 0, 64),

		UpLeft:    make([]BorderLine, 0, 64),
		UpRight:   make([]BorderLine, 0, 64),
		DownLeft:  make([]BorderLine, 0, 64),
		DownRight: make([]BorderLine, 0, 64),
	}

	// We do not accept borders in the outer ring. The terrain must therefore be enclosed by a shell of solid (non-diagonal) blocks.
	// This shell must not neccessarily be the outer ring.

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
			if HasBorderTowards(mine, above, UP) && x != width-1 {
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
			if HasBorderTowards(above, mine, DOWN) && x != width-1 {
				if downwardsBorderStart == -1 {
					downwardsBorderStart = x // the border just started
				}
			} else {
				if downwardsBorderStart != -1 { // the border just ended
					downwardsBorderEnd := x
					borders.Left = append(borders.Left, BorderLine{ // above = solid
						StartX: downwardsBorderEnd, // the border goes from right to left (solid region must be on the  right side)
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
			if HasBorderTowards(mine, left, LEFT) && y != height-1 {
				if leftBorderStart == -1 {
					leftBorderStart = y // the border just started
				}
			} else {
				if leftBorderStart != -1 { // the border just ended
					leftBorderEnd := y
					borders.Up = append(borders.Up, BorderLine{ // right = solid
						StartX: x,
						StartY: leftBorderEnd,
						Length: leftBorderEnd - leftBorderStart,
					})
					leftBorderStart = -1
				}
			}

			// Border facing to the right
			if HasBorderTowards(left, mine, RIGHT) && y != height-1 {
				if rightBorderStart == -1 {
					rightBorderStart = y // the border just started
				}
			} else {
				if rightBorderStart != -1 { // the border just ended
					rightBorderEnd := y
					borders.Down = append(borders.Down, BorderLine{ // left = solid
						StartX: x,
						StartY: rightBorderStart, // the border goes from right to left (solid region must be on the right side)
						Length: rightBorderEnd - rightBorderStart,
					})
					rightBorderStart = -1
				}
			}
		}
	}

	// Find diagonal borders from the top-left to the bottom-right:
	diagonalChecks := width + height - 1
	// For diagonal tiles, we do not ignore the outer ring. But if we find diagonals there, we emmit an error
	for d := 0; d < diagonalChecks; d++ {
		var firstX int
		var firstY int

		if d < width {
			firstX = d
			firstY = 0
		} else {
			firstX = 0
			firstY = d - width + 1
		}

		upRightBorderStart := -1
		downLeftBorderStart := -1

		x := firstX
		y := firstY
		for i := 0; ; i++ {
			var tile Tile
			if tile, err = layer.GetTile(x, y, width, height); err != nil {
				return borders, fmt.Errorf("Failed to compute diagonal border (%dx%d): %v", x, y, err)
			}

			// border facing up-right
			if tile.GetType() == SOLID_AT_LOWER_LEFT {
				if x == 0 || y == 0 || x == width-1 || y == height-1 {
					log.Warningf("The outer ring of the map contains diagonal tiles. Note that the whole area that is reachable within the game must be enclosed by solid, non-diagonal tiles. Position: %vx%v", x, y)
				}
				if upRightBorderStart == -1 {
					upRightBorderStart = i // the border just started
				}
			} else {
				if upRightBorderStart != -1 { // the border just ended
					borders.DownRight = append(borders.DownRight, BorderLine{ // bottom left = solid, border pointing down-right
						StartX: firstX + upRightBorderStart,
						StartY: firstY + upRightBorderStart,
						Length: i - upRightBorderStart,
					})
					upRightBorderStart = -1
				}
			}

			// border facing down-left
			if tile.GetType() == SOLID_AT_UPPER_RIGHT {
				if x == 0 || y == 0 || x == width-1 || y == height-1 {
					log.Warningf("The outer ring of the map contains diagonal tiles. Note that the whole area that is reachable within the game must be enclosed by solid, non-diagonal tiles. Position: %vx%v", x, y)
				}
				if downLeftBorderStart == -1 {
					downLeftBorderStart = i // the border just started
				}
			} else {
				if downLeftBorderStart != -1 { // the border just ended
					borders.UpLeft = append(borders.UpLeft, BorderLine{ // upper right = solid, border pointing up-left
						StartX: firstX + i,
						StartY: firstY + i, // The border goes from down-right to upper-left
						Length: i - downLeftBorderStart,
					})
					downLeftBorderStart = -1
				}
			}
			x++
			y++
			if x >= width || y >= height {
				break
			}
		}
	}

	// Find diagonal borders from the bottom-left to the top-right:
	for d := 0; d < diagonalChecks; d++ {
		var firstX int
		var firstY int

		if d < width {
			firstX = d
			firstY = height - 1
		} else {
			firstX = 0
			firstY = d - width
		}

		upLeftBorderStart := -1
		downRightBorderStart := -1

		x := firstX
		y := firstY
		for i := 0; ; i++ {
			var tile Tile
			if tile, err = layer.GetTile(x, y, width, height); err != nil {
				return borders, fmt.Errorf("Failed to compute diagonal border (%dx%d): %v", x, y, err)
			}

			// border facing up-left
			if tile.GetType() == SOLID_AT_LOWER_RIGHT {
				if x == 0 || y == 0 || x == width-1 || y == height-1 {
					log.Warningf("The outer ring of the map contains diagonal tiles. Note that the whole area that is reachable within the game must be enclosed by solid, non-diagonal tiles. Position: %vx%v", x, y)
				}
				if upLeftBorderStart == -1 {
					upLeftBorderStart = i // the border just started
				}
			} else {
				if upLeftBorderStart != -1 { // the border just ended
					borders.UpRight = append(borders.UpRight, BorderLine{ // bottom right = solid, border pointing up-right
						StartX: firstX + upLeftBorderStart,
						StartY: firstY - upLeftBorderStart + 1,
						Length: i - upLeftBorderStart,
					})
					upLeftBorderStart = -1
				}
			}

			// border facing down-right
			if tile.GetType() == SOLID_AT_UPPER_LEFT {
				if x == 0 || y == 0 || x == width-1 || y == height-1 {
					log.Warningf("The outer ring of the map contains diagonal tiles. Note that the whole area that is reachable within the game must be enclosed by solid, non-diagonal tiles. Position: %vx%v", x, y)
				}
				if downRightBorderStart == -1 {
					downRightBorderStart = i // the border just started
				}
			} else {
				if downRightBorderStart != -1 { // the border just ended
					borders.DownLeft = append(borders.DownLeft, BorderLine{ // upper left = solid, border pointing down-left
						StartX: firstX + i,
						StartY: firstY - i + 1, // The border goes from down-right to upper-left
						Length: i - downRightBorderStart,
					})
					downRightBorderStart = -1
				}
			}
			x++
			y--
			if x >= width || y < 0 {
				break
			}
		}
	}

	// Possible optimisation: if the map contains unreachable positions, it's borders can be dropped

	// Validate and reduce:
	// if len(borders.Left) == 0 || len(borders.Right) == 0 || len(borders.Up) == 0 || len(borders.Down) == 0 ||
	//     len(borders.UpLeft) == 0 || len(borders.UpRight) == 0 || len(borders.DownLeft) == 0 || len(borders.DownRight) == 0 {
	//     return borders, fmt.Errorf("Invalid map: Failed to compute border. A closed map contains at least one border in each direction. "+
	//         "Found (left, right, up, down): %d, %d, %d, %d "+
	//         "Found (up-left, up-right, down-left, down-right): %d, %d, %d, %d ",
	//         len(borders.Left), len(borders.Right), len(borders.Up), len(borders.Down),
	//         len(borders.UpLeft), len(borders.UpRight), len(borders.DownLeft), len(borders.DownRight))
	// }

	return borders, nil
}
