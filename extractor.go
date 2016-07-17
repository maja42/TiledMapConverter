package main

import (
    "fmt"
)

type Player struct {
    SpawnX            int
    SpawnY            int
    BaseBuildingFlags uint8 // needed for rotation
    Units             []Unit
}

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

type UnitMapping struct {
    // []UnitMapping: tile-index to player&unit type
    Player int
    Type   UnitType
}

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

func GetTileMapping() (map[uint32]BaseMapping, map[uint32]UnitMapping) {
    basemapping := make(map[uint32]BaseMapping)
    unitmapping := make(map[uint32]UnitMapping)

    // bases
    basemapping[82] = BaseMapping{0}  //Base of player 1
    basemapping[104] = BaseMapping{1} //Base of player 2

    // Offense units
    unitmapping[60] = UnitMapping{0, UnitType_Offense}
    unitmapping[58] = UnitMapping{1, UnitType_Offense}
    return basemapping, unitmapping
}

func ExtractSpawnInfo(tilemap TileMap) ([]Player, error) {
    var spawnLayerIdx int = -1
    for idx, layer := range tilemap.Layers {
        if layer.Name != "spawn" {
            continue
        }
        if spawnLayerIdx == -1 {
            spawnLayerIdx = idx
        } else {
            return nil, fmt.Errorf("Multiple layers with name 'spawn' found")
        }
    }
    if spawnLayerIdx == -1 {
        return nil, fmt.Errorf("No layer with name 'spawn' found")
    }

    player, err := ExtractSpawnInfoFromLayer(tilemap.Width, tilemap.Height, tilemap.Layers[spawnLayerIdx])
    if err != nil {
        return nil, err
    }
    tilemap.Layers = append(tilemap.Layers[:spawnLayerIdx], tilemap.Layers[spawnLayerIdx+1:]...)
    return player, nil
}

func ExtractSpawnInfoFromLayer(width, height int, layer TileMapLayer) ([]Player, error) {
    var players = make([]Player, 8)
    for i := 0; i < 8; i++ {
        players[i] = *NewPlayer()
    }

    basemapping, unitmapping := GetTileMapping()

    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            idx := y*width + x
            tile := layer.Tiles[idx].Index
            flags := layer.Tiles[idx].Flags

            // check if this is a base tile
            {
                mapping, ok := basemapping[tile]
                if ok {
                    if mapping.Player < 0 || mapping.Player >= 8 || players[mapping.Player].SpawnX != -1 {
                        return nil, fmt.Errorf("Failed to map tile: Invalid base building mapping or multiple base buildings for player %d (Tile = %d)", mapping.Player, tile)
                    }

                    players[mapping.Player].SpawnX = x
                    players[mapping.Player].SpawnY = y
                    players[mapping.Player].BaseBuildingFlags = flags
                    continue
                }
            }

            // check if this is a unit tile
            {
                mapping, ok := unitmapping[tile]
                if ok {
                    if mapping.Player < 0 || mapping.Player >= 8 {
                        return nil, fmt.Errorf("Failed to map tile: Invalid unit mapping for player %d (Tile = %d)", mapping.Player, tile)
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
    var actualPlayers = make([]Player, 0)
    for i, p := range players {
        if p.SpawnX < 0 || p.SpawnY < 0 { // Player does not exist
            if len(p.Units) != 0 {
                return nil, fmt.Errorf("Invalid map: Player %d has no base building, but has units.", i)
            }
            continue
        }
        actualPlayers = append(actualPlayers, p)
    }
    if len(actualPlayers) <= 1 {
        return nil, fmt.Errorf("Invalid map: Does not contain enough player spawn points. (Needed >=2, Found %d)", len(actualPlayers))
    }

    return actualPlayers, nil
}
