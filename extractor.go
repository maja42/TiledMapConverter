package main

import (
	"fmt"
)

// ResourcePoint contains all information about the spawn of a single resource-point.
type ResourcePoint struct {
	SpawnX             int
	SpawnY             int
	ResourcePointFlags uint8 // needed for rotation
}

// WaterdropSource contains all information about the spawn of a water drop source that continuously spawns drops falling of the roof.
type WaterdropSource struct {
	SpawnX         int
	SpawnY         int
	WaterdropFlags uint8
}

// Player contains all spawn inform about a single player in the game.
type Player struct {
	Buildings []Building
	Units     []Unit
}

// Unit contains all spawn information about a unit that should spawn at game start.
type Unit struct {
	Type   UnitType
	SpawnX int
	SpawnY int
}

type UnitType int

const (
	UnitType_Offense      UnitType = 1
	UnitType_Defense      UnitType = 2
	UnitType_LongRange    UnitType = 3
	UnitType_Special      UnitType = 4
	UnitType_Construction UnitType = 5
)

// UnitMapping defines which .tmx tiles (tile-index) are used to spawn a unit
type UnitMapping struct {
	// []UnitMapping: tile-index to player&unit type.
	Player int
	Type   UnitType
}

type Building struct {
	Type   BuildingType
	SpawnX int
	SpawnY int
	Flags  uint8 // needed for rotation
}

type BuildingType int

const (
	BuildingType_Base    BuildingType = 1
	BuildingType_Pump    BuildingType = 2
	BuildingType_Factory BuildingType = 3
	BuildingType_Turret  BuildingType = 4
	BuildingType_Bridge  BuildingType = 5
)

// BuildingMapping defines which .tmx tiles (tile-index) are used to spawn a building
type BuildingMapping struct {
	// []BuildingMapping: tile-index to building type.
	Type BuildingType
}

// PlayerMapping defines which .tmx tiles (tile-index) are used to spawn a building of a specific player (each building has a player-token in the up-left corner)
type PlayerMapping struct {
	// tile-index to player
	Player int
}

func NewPlayer() *Player {
	return &Player{
		Buildings: make([]Building, 0),
		Units:     make([]Unit, 0),
	}
}

func GetTileMapping() (uint32, uint32, map[uint32]PlayerMapping, map[uint32]BuildingMapping, map[uint32]UnitMapping) {
	playermapping := make(map[uint32]PlayerMapping)
	buildingmapping := make(map[uint32]BuildingMapping)
	unitmapping := make(map[uint32]UnitMapping)

	// resource spawn mapping
	var resourceMapping uint32 = 173

	// water drop mapping
	var waterdropSpawnMapping uint32 = 177

	// Unit + Player mapping
	for i := 0; i < 8; i++ {
		var firstIdx = uint32(1 + i*10 + (i/2)*20)

		unitmapping[firstIdx+0] = UnitMapping{i, UnitType_Offense}
		unitmapping[firstIdx+2] = UnitMapping{i, UnitType_Defense}
		unitmapping[firstIdx+4] = UnitMapping{i, UnitType_LongRange}
		unitmapping[firstIdx+6] = UnitMapping{i, UnitType_Special}
		unitmapping[firstIdx+8] = UnitMapping{i, UnitType_Construction}
		playermapping[firstIdx+9] = PlayerMapping{i}
	}

	// Building mapping
	// For buildings, the upper-left tile is the player-token (playermapping). The tile on the right (depends on the rotation) defines the building type. So 2 tiles are responsible for defining a building.
	buildingmapping[162] = BuildingMapping{BuildingType_Base}
	buildingmapping[234] = BuildingMapping{BuildingType_Pump}
	buildingmapping[238] = BuildingMapping{BuildingType_Turret}

	return resourceMapping, waterdropSpawnMapping, playermapping, buildingmapping, unitmapping
}

func ExtractSpawnInfo(tilemap *TileMap) ([]ResourcePoint, []WaterdropSource, []Player, error) {
	spawnLayerIdx, err := tilemap.GetLayer("spawn")
	if err != nil {
		return nil, nil, nil, err
	}

	resources, waterdropSources, player, err := ExtractSpawnInfoFromLayer(tilemap.Width, tilemap.Height, &tilemap.Layers[spawnLayerIdx])
	if err != nil {
		return nil, nil, nil, err
	}
	tilemap.Layers = append(tilemap.Layers[:spawnLayerIdx], tilemap.Layers[spawnLayerIdx+1:]...) // remove spawn layer from tilemap
	return resources, waterdropSources, player, nil
}

func ExtractSpawnInfoFromLayer(width, height int, layer *TileMapLayer) ([]ResourcePoint, []WaterdropSource, []Player, error) {
	var players = make([]Player, 8)
	for i := 0; i < 8; i++ {
		players[i] = *NewPlayer()
	}

	var resources = make([]ResourcePoint, 0, 16)
	var waterdrops = make([]WaterdropSource, 0, 4)

	resourceMapping, waterdropSpawnMapping, playerMapping, buildingMapping, unitMapping := GetTileMapping()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			tile := layer.Tiles[idx]

			if tile.Index != 0 {
				if tile.TileSet == nil {
					return nil, nil, nil, fmt.Errorf("Invalid map: Unknown tileset (x=%d, y=%d, layer=%q)", x, y, layer.Name)
				} else if tile.TileSet.Type != SPAWN_TILESET {
					return nil, nil, nil, fmt.Errorf("Invalid tileset: The tile (x=%d, y=%d, layer=%q) should be part of the Spawn TileSet, but it is part of the tileset %q.", x, y, layer.Name, tile.TileSet.Name)
				}
			}

			tileID := tile.Index
			flags := tile.Flags

			// check if this is a resource spawn tile
			{
				if tileID == resourceMapping {
					if tile.IsMirrored() {
						return nil, nil, nil, fmt.Errorf("Failed to map tile: Resource points must not be mirrored, only rotations are allowed.  (x=%d, y=%d)", x, y)
					}
					resources = append(resources, ResourcePoint{
						SpawnX:             x,
						SpawnY:             y,
						ResourcePointFlags: flags,
					})
				}
			}

			// check if this is a water drop spawn tile
			{
				if tileID == waterdropSpawnMapping {
					waterdrops = append(waterdrops, WaterdropSource{
						SpawnX:         x,
						SpawnY:         y,
						WaterdropFlags: flags,
					})
				}
			}

			// check if this is a unit tile
			{
				mapping, ok := unitMapping[tileID]
				if ok {
					if mapping.Player < 0 || mapping.Player >= 8 {
						return nil, nil, nil, fmt.Errorf("Failed to map tile: Invalid unit mapping for player %d (Tile = %d)", mapping.Player, tileID)
					}
					if flags != 0 {
						return nil, nil, nil, fmt.Errorf("Failed to map tile: Units must not be mirrored or rotated. (player %d, x=%d, y=%d, layer=%q)", mapping.Player, x, y, layer.Name)
					}

					newUnit := Unit{
						Type:   mapping.Type,
						SpawnX: x,
						SpawnY: y,
					}
					players[mapping.Player].Units = append(players[mapping.Player].Units, newUnit)
					continue
				}
			}

			// check if this is a building tile
			{
				mapping, ok := playerMapping[tileID]
				if ok {
					if mapping.Player < 0 || mapping.Player >= 8 {
						return nil, nil, nil, fmt.Errorf("Failed to map tile: Invalid player mapping for player %d (Tile = %d, x=%d, y=%d, layer=%q)", mapping.Player, tileID, x, y, layer.Name)
					}
					if tile.IsMirrored() {
						return nil, nil, nil, fmt.Errorf("Failed to map tile: Buildings must not be mirrored, only rotations are allowed. The player mapping tile (x=%d, y=%d, layer=%q) is mirrored", x, y, layer.Name)
					}

					// Now we know which player this building belongs to and how it is oriented. Now we need to know which type of building this is
					var newBuilding Building
					newBuilding.SpawnX = x
					newBuilding.SpawnY = y
					newBuilding.Flags = flags

					vecX, vecY := tile.GetRightVector()
					identX, identY := x+vecX, y+vecY
					buildingTile := layer.Tiles[identY*width+identX]

					if buildingTile.TileSet == nil {
						return nil, nil, nil, fmt.Errorf("Invalid map: Unknown tileset. The tile (x=%d, y=%d, layer=%q) should be part of the Spawn TileSet, but is empty.", identX, identY, layer.Name)
					} else if tile.TileSet.Type != SPAWN_TILESET {
						return nil, nil, nil, fmt.Errorf("Invalid tileset: The tile (x=%d, y=%d, layer=%q) should be part of the Spawn TileSet, but it is part of the tileset %q.", identX, identY, layer.Name, tile.TileSet.Name)
					}

					tileID := buildingTile.Index
					buildingFlags := buildingTile.Flags
					if buildingFlags != flags {
						return nil, nil, nil, fmt.Errorf("Invalid map: Inconsistent tile flags. The player mapping tile (x=%d, y=%d) and building tile (x=%d, y=%d) must have the same flags (layer=%q).", x, y, identX, identY, layer.Name)
					}

					buildingMapping, ok := buildingMapping[tileID]
					if !ok {
						return nil, nil, nil, fmt.Errorf("Invalid map: There exists a player-mapping tile (x=%d, y=%d) which indicates that there should be a building-spawn. However, the tile (x=%d, y=%d) has no valid building-mapping tile (layer=%q).", x, y, identX, identY, layer.Name)
					}

					newBuilding.Type = buildingMapping.Type
					players[mapping.Player].Buildings = append(players[mapping.Player].Buildings, newBuilding)
					continue
				}
			}

		}
	}

	// Validate and reduce:
	if len(resources) < 1 {
		return nil, nil, nil, fmt.Errorf("Invalid map: Does not contain any resource points. (Needs >=1, Found %d)", len(resources))
	}
	var actualPlayers = make([]Player, 0)
	for i, p := range players {
		baseBuildingCount := 0
		for _, b := range p.Buildings {
			if b.Type == BuildingType_Base {
				baseBuildingCount++
			}
		}

		if baseBuildingCount <= 0 { // Player does not exist
			if len(p.Units) != 0 {
				return nil, nil, nil, fmt.Errorf("Invalid map: Player %d has no base building, but has units.", i)
			}
			if len(p.Buildings) != 0 {
				return nil, nil, nil, fmt.Errorf("Invalid map: Player %d has no base building, but has other buildings.", i)
			}
			continue
		}
		if baseBuildingCount > 1 {
			log.Warningf("Warning: Player %d has %d base buildings (more than one). This is ok, but maybe not intended.", i, baseBuildingCount)
		}
		actualPlayers = append(actualPlayers, p)
	}
	if len(actualPlayers) <= 1 {
		return nil, nil, nil, fmt.Errorf("Invalid map: Does not contain enough player spawn points. (Needed >=2, Found %d)", len(actualPlayers))
	}

	return resources, waterdrops, actualPlayers, nil
}
