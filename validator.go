package main

import (
    "fmt"
)

func ValidateTileMap(tilemap TileMap) error {
    if tilemap.Version != "1.0" {
        log.Warningf("The tiles file was stored with an unsupported version: '%s'", tilemap.Version)
    }
    if tilemap.Width <= 0 {
        return fmt.Errorf("Invalid tilemap width: %d", tilemap.Width)
    }
    if tilemap.Height <= 0 {
        return fmt.Errorf("Invalid tilemap height: %d", tilemap.Height)
    }
    if tilemap.Orientation != "orthogonal" {
        return fmt.Errorf("Invalid orientation: '%s'", tilemap.Orientation)
    }
    if tilemap.Renderorder != "right-down" {
        return fmt.Errorf("Invalid render order: '%s'", tilemap.Renderorder)
    }
    if tilemap.Tilewidth != 256 || tilemap.Tileheight != 256 {
        return fmt.Errorf("Invalid tile size: %dx%d", tilemap.Tilewidth, tilemap.Tileheight)
    }
    if len(tilemap.Layers) <= 0 && len(tilemap.Layers) >= 256 {
        return fmt.Errorf("Invalid layer count: %d", len(tilemap.Layers))
    }
    if (len(tilemap.Tilesets) <= 0) {
        return fmt.Errorf("No tileset detected.");
    }
    return nil
}
