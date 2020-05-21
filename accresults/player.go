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
}

func (player *Player) mergeDriver(driver *Driver) {
    newName := &Name{driver.FirstName, driver.LastName, driver.ShortName}
    player.MostRecentName = newName
    var exists = false
    for _, name := range player.Names {
        exists = exists || reflect.DeepEqual(name, newName)
    }
    if !exists {
        player.Names = append(player.Names, newName)
    }
}
