package main

import (
    "fmt"
)

// ResourcePoint contains all information about the spawn-position of a single resource-point.
type ResourcePoint struct {
    SpawnX             int
    SpawnY             int
    ResourcePointFlags uint8 // needed for rotation
}

// Player contains all spawn inform about a single player in the game.
type Player struct {
    SpawnX            int
    SpawnY            int
    BaseBuildingFlags uint8 // needed for rotation
    Units             []Unit
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

// BaseMapping defines which .tmx tiles (tile-index) are used to spawn a player base
type BaseMapping struct {
    // []BaseMapping: tile-index to player
    Player int
}

func NewPlayer() *Player {
    return &Player{
        SpawnX:            -1,
        SpawnY:            -1,
        BaseBuildingFlags: 0x00,
        Units:             make([]Unit, 0),
    }
}

func GetTileMapping() (uint32, map[uint32]BaseMapping, map[uint32]UnitMapping) {
    basemapping := make(map[uint32]BaseMapping)
    unitmapping := make(map[uint32]UnitMapping)

    // resource spawns
    var resourcemapping uint32 = 14

    // bases
    basemapping[82] = BaseMapping{0}  //Base of player 1
    basemapping[104] = BaseMapping{1} //Base of player 2

    // Offense units
    unitmapping[60] = UnitMapping{0, UnitType_Offense}
    unitmapping[58] = UnitMapping{1, UnitType_Offense}
    return resourcemapping, basemapping, unitmapping
}

func ExtractSpawnInfo(tilemap TileMap) ([]ResourcePoint, []Player, error) {
    spawnLayerIdx, err := tilemap.GetLayer("spawn")
    if err != nil {
        return nil, nil, err
    }

    resources, player, err := ExtractSpawnInfoFromLayer(tilemap.Width, tilemap.Height, tilemap.Layers[spawnLayerIdx])
    if err != nil {
        return nil, nil, err
    }
    tilemap.Layers = append(tilemap.Layers[:spawnLayerIdx], tilemap.Layers[spawnLayerIdx+1:]...) // remove spawn layer from tilemap
    return resources, player, nil
}

func ExtractSpawnInfoFromLayer(width, height int, layer TileMapLayer) ([]ResourcePoint, []Player, error) {
    var players = make([]Player, 8)
    for i := 0; i < 8; i++ {
        players[i] = *NewPlayer()
    }

    var resources = make([]ResourcePoint, 0, 16)

    resourcePointMapping, basemapping, unitmapping := GetTileMapping()

    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            idx := y*width + x
            tile := layer.Tiles[idx]

            var offset uint32
            if (tile.TileSet == nil) {
                offset = 0
            } else {
                offset = tile.TileSet.FirstGid - 1
            }

            tileID := tile.Index - offset
            flags := tile.Flags

            // check if this is a resource spawn tile
            {
                if tileID == resourcePointMapping {
                    resources = append(resources, ResourcePoint{
                        SpawnX:             x,
                        SpawnY:             y,
                        ResourcePointFlags: flags,
                    })
                }
            }

            // check if this is a base tile
            {
                mapping, ok := basemapping[tileID]
                if ok {
                    if mapping.Player < 0 || mapping.Player >= 8 || players[mapping.Player].SpawnX != -1 {
                        return nil, nil, fmt.Errorf("Failed to map tile: Invalid base building mapping or multiple base buildings for player %d (Tile = %d)", mapping.Player, tileID)
                    }

                    players[mapping.Player].SpawnX = x
                    players[mapping.Player].SpawnY = y
                    players[mapping.Player].BaseBuildingFlags = flags
                    continue
                }
            }

            // check if this is a unit tile
            {
                mapping, ok := unitmapping[tileID]
                if ok {
                    if mapping.Player < 0 || mapping.Player >= 8 {
                        return nil, nil, fmt.Errorf("Failed to map tile: Invalid unit mapping for player %d (Tile = %d)", mapping.Player, tileID)
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

        }
    }

    // Validate and reduce:
    if len(resources) < 1 {
        return nil, nil, fmt.Errorf("Invalid map: Does not contain any resource points. (Needed >=1, Found %d)", len(resources))
    }
    var actualPlayers = make([]Player, 0)
    for i, p := range players {
        if p.SpawnX < 0 || p.SpawnY < 0 { // Player does not exist
            if len(p.Units) != 0 {
                return nil, nil, fmt.Errorf("Invalid map: Player %d has no base building, but has units.", i)
            }
            continue
        }
        actualPlayers = append(actualPlayers, p)
    }
    if len(actualPlayers) <= 1 {
        return nil, nil, fmt.Errorf("Invalid map: Does not contain enough player spawn points. (Needed >=2, Found %d)", len(actualPlayers))
    }

    return resources, actualPlayers, nil
}
