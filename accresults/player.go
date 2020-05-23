package accresults

import (
    "reflect"
)

type Name struct {
    FirstName string
    LastName string
    ShortName string
}

type Player struct {
    PlayerId string
    MostRecentName *Name
    Names []*Name
    SessionNames []string
    Events map[string]*Event
}

func NewPlayer(playerId string) *Player {
    return &Player{
        playerId,
        nil,
        nil,
        nil,
        make(map[string]*Event),
    }
}

func (player *Player) mergeDriver(driver *Driver) {
    newName := &Name{driver.FirstName, driver.LastName, driver.ShortName}
    player.MostRecentName = newName
    
    // PlayerID S76561197960267366 is shared between multiple players
    if player.PlayerId == "S76561197960267366" {
        player.MostRecentName = &Name{"Unknown", "Driver", "-"}
    }
    
    var exists = false
    for _, name := range player.Names {
        exists = exists || reflect.DeepEqual(name, newName)
    }
    if !exists {
        player.Names = append(player.Names, newName)
    }
}
