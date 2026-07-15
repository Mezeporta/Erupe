package channelserver

import (
	"math/rand"
	"slices"
)

// Tile is a single node in a guild interception map's tile graph.
type Tile struct {
	ID          uint16
	NextID      uint16
	BranchID    uint16
	QuestFile1  uint16
	QuestFile2  uint16
	QuestFile3  uint16
	BranchIndex uint8
	Type        uint8
	PointsReq   int32
	Claimed     bool
	Unk1        uint8
	Unk2        uint32
}

// MapData is one of the guild's interception maps (main path + branch tiles).
type MapData struct {
	ID     uint32
	NextID uint32
	Points map[uint16]int32
	Tiles  []Tile
}

// GetClaimed returns the number of tiles claimable with the map's currently
// accumulated points, deducting each claimed tile's cost along the way.
func (md *MapData) GetClaimed() uint32 {
	var claimed uint32
	for _, tile := range md.Tiles {
		if md.Points[tile.QuestFile1]-tile.PointsReq > 0 {
			tile.Claimed = true
			if tile.PointsReq > 0 {
				claimed++
			}
			md.Points[tile.QuestFile1] -= tile.PointsReq
		}
	}
	return claimed
}

// TotalPoints sums the point cost of the map's main-path tiles (excludes
// branch/treasure tiles, type > 2).
func (md *MapData) TotalPoints() int32 {
	var points int32
	for i := range md.Tiles {
		if md.Tiles[i].Type > 2 {
			continue
		}
		points += md.Tiles[i].PointsReq
	}
	return points
}

// Completed reports whether the map's main path has been fully claimed.
func (md *MapData) Completed() bool {
	return md.Points[0] > md.TotalPoints()
}

// MapBranch describes a treasure chest reward branching off a map's main path.
type MapBranch struct {
	MapIndex   uint32
	ItemType   uint8
	ItemID     uint16
	Quantity   uint16
	TileIndex1 uint16 // Sequential
	TileIndex2 uint16 // Sequential, last = 99
	ChestType  uint8
}

// InterceptionMaps is a guild's full set of interception maps, persisted as
// JSON in the guilds.interception_maps column.
type InterceptionMaps struct {
	Maps     []MapData
	Branches []MapBranch
}

// CurrPrevID returns the ID of the map currently in progress and the one
// before it (0 if there is no previous map).
func (im *InterceptionMaps) CurrPrevID() (uint32, uint32) {
	var currID, prevID uint32
	for i := range im.Maps {
		prevID = currID
		currID = im.Maps[i].ID
		if im.Maps[i].Points[0] < im.Maps[i].TotalPoints() {
			break
		}
	}
	return currID, prevID
}

// getNeighbourTiles returns the IDs of tiles adjacent to tile on the
// hex-like 5x12 grid encoded as (row+1)*100 + col+1.
func getNeighbourTiles(tiles [][]uint16, tile uint16) []uint16 {
	var vals []uint16
	var temp []uint16
	if tile%2 == 0 {
		temp = []uint16{tile - 100, tile - 1, tile + 1, tile + 99, tile + 100, tile + 101}
	} else {
		temp = []uint16{tile - 101, tile - 100, tile - 99, tile - 1, tile + 1, tile + 100}
	}

	for _, val := range temp {
		for x := range tiles {
			for y := range tiles[x] {
				if tiles[x][y] == val {
					vals = append(vals, val)
				}
			}
		}
	}
	return vals
}

// getBranchTile picks a random valid neighbour of tile to extend a branch
// path into, avoiding tiles already used by excluded paths.
func getBranchTile(tiles [][]uint16, excluded []uint16, tile uint16) uint16 {
	neighbours := getNeighbourTiles(tiles, tile)
	var validNeighbours, validBranchTiles []uint16
	for i := range neighbours {
		if !slices.Contains(excluded, neighbours[i]) {
			// Neighbour tiles that are not in the path
			validNeighbours = append(validNeighbours, neighbours[i])
		}
	}
	if len(validNeighbours) == 0 {
		return 0
	}
	for i := range validNeighbours {
		subNeighbours := getNeighbourTiles(tiles, validNeighbours[i])
		var invalid bool
		var cleanSubNeighbours []uint16
		for _, subNeighbour := range subNeighbours {
			if subNeighbour != validNeighbours[i] && subNeighbour != tile {
				cleanSubNeighbours = append(cleanSubNeighbours, subNeighbour)
			}
		}
		for _, subNeighbour := range cleanSubNeighbours {
			if slices.Contains(excluded, subNeighbour) {
				invalid = true
				break
			}
		}
		if !invalid {
			validBranchTiles = append(validBranchTiles, validNeighbours[i])
		}
	}
	if len(validBranchTiles) == 0 {
		return 0
	}
	return validBranchTiles[rand.Intn(len(validBranchTiles))]
}

// GenerateUdGuildMaps procedurally generates a fresh set of 5 guild
// interception maps (main path + branch treasure tiles), each harder and
// worth more points than the last.
func GenerateUdGuildMaps() ([]MapData, []MapBranch) {
	tiles := make([][]uint16, 5)
	for i := range tiles {
		tiles[i] = make([]uint16, 12)
		for j := range tiles[i] {
			tiles[i][j] = uint16(((i + 1) * 100) + j + 1)
		}
	}

	var mapData []MapData
	var mapBranches []MapBranch

	for i := 0; i < 5; i++ {
		var startTile, endTile uint16
		var randTemp []uint16
		randTemp = tiles[rand.Intn(len(tiles))]
		startTile = randTemp[rand.Intn(len(randTemp))]
		for {
			randTemp = tiles[rand.Intn(len(tiles))]
			endTile = randTemp[rand.Intn(len(randTemp))]
			invalidTiles := append(getNeighbourTiles(tiles, startTile), startTile)
			if !slices.Contains(invalidTiles, endTile) {
				break
			}
		}

		var tilePath []uint16
		var iterations int
		var tooDifficult bool
		for {
			var pathFailed bool
			var evictedTiles []uint16
			tilePath = []uint16{startTile}
			for {
				var possibleTiles []uint16
				tempTiles := getNeighbourTiles(tiles, tilePath[len(tilePath)-1])
				for _, tile := range tempTiles {
					if !slices.Contains(evictedTiles, tile) {
						possibleTiles = append(possibleTiles, tile)
					}
				}
				if len(possibleTiles) == 0 {
					pathFailed = true
					break
				}
				evictedTiles = append(evictedTiles, possibleTiles...)
				newTile := possibleTiles[rand.Intn(len(possibleTiles))]
				tilePath = append(tilePath, newTile)
				if tilePath[len(tilePath)-1] == endTile {
					if len(tilePath) < 20 {
						pathFailed = true
					}
					break
				}
			}
			if !pathFailed {
				break
			}
			iterations++
			if iterations > 1000 {
				tooDifficult = true
				break
			}
		}

		if tooDifficult {
			i--
			continue
		}

		var mapTiles []Tile
		for j, tile := range tilePath {
			mapTile := Tile{}
			mapTile.ID = tile
			mapTile.BranchIndex = uint8(j + 1)
			switch j {
			case 0:
				mapTile.Type = 1
				mapTile.NextID = tilePath[j+1]
			case len(tilePath) - 1:
				mapTile.Type = 2
			default:
				mapTile.NextID = tilePath[j+1]
			}
			switch i {
			case 0:
				mapTile.PointsReq = int32(2500 + 150*(j-1))
			case 1:
				mapTile.PointsReq = int32(5500 + 600*(j-1))
			case 2:
				mapTile.PointsReq = int32(6500 + 800*(j-1))
			case 3:
				mapTile.PointsReq = int32(7500 + 1000*(j-1))
			case 4:
				mapTile.PointsReq = int32(8500 + 1000*(j-1))
			}
			if mapTile.Type == 1 {
				mapTile.PointsReq = 0
			}
			mapTiles = append(mapTiles, mapTile)
		}

		evictedTiles := append([]uint16{}, tilePath...)

		var branchTiles []Tile
		for j := range mapTiles {
			if mapTiles[j].Type != 0 {
				continue
			}
			var newBranchTile uint16
			var branchIndex int
			currentBranchTile := mapTiles[j]
			for {
				newBranchTile = getBranchTile(tiles, evictedTiles, currentBranchTile.ID)
				if newBranchTile == 0 {
					if currentBranchTile != mapTiles[j] {
						branchTiles[len(branchTiles)-1].Type = 4
						branchTiles[len(branchTiles)-1].Unk1 = 1
						branchTiles[len(branchTiles)-1].Unk2 = 2
						// Make treasure more interesting, 2000GCP for now
						mapBranches = append(mapBranches, MapBranch{
							MapIndex:   uint32(i + 1),
							ItemType:   26,
							ItemID:     0,
							Quantity:   2000,
							TileIndex1: uint16(branchIndex),
							TileIndex2: 99,
							ChestType:  2,
						})
					}
					break
				}
				if currentBranchTile.ID == mapTiles[j].ID {
					mapTiles[j].BranchID = newBranchTile
					mapTiles[j].Type = 3
				} else {
					branchTiles[len(branchTiles)-1].NextID = newBranchTile
				}
				branchIndex++
				newTile := Tile{
					ID:          newBranchTile,
					QuestFile1:  uint16(j%5 + 58079),
					BranchIndex: uint8(branchIndex),
					Type:        0,
					PointsReq:   100,
					Unk1:        0,
					Unk2:        0,
				}
				branchTiles = append(branchTiles, newTile)
				evictedTiles = append(evictedTiles, getNeighbourTiles(tiles, currentBranchTile.ID)...)
				currentBranchTile = newTile
			}
		}
		mapTiles = append(mapTiles, branchTiles...)
		if i >= 4 {
			mapData = append(mapData, MapData{uint32(i + 1), 4, make(map[uint16]int32), mapTiles})
		} else {
			mapData = append(mapData, MapData{uint32(i + 1), uint32(i + 2), make(map[uint16]int32), mapTiles})
		}
	}
	return mapData, mapBranches
}
