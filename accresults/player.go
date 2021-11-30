package accresults

import (
	"math"
	"reflect"

	"github.com/geniusdex/racce/accdata"
)

type Name struct {
	FirstName string
	LastName  string
	ShortName string
}

// PlayerTrackCarData contains information of a player on a track with a car model
type PlayerTrackCarData struct {
	// Player points to the player that this data is about
	Player *Player
	// TrackName is the track label that this data is about
	TrackName string
	// CarModel is the car model that this data is about
	CarModel int

	// BestLap is the fastest laptime achieved by this player on this track in this car
	BestLap int
}

// PlayerTrackData contains information of a player on a track
type PlayerTrackData struct {
	// Player points to the player that this data is about
	Player *Player
	// TrackName is the track label that this data is about
	TrackName string

	// BestLap is the fastest laptime achieved by this player on this track
	BestLap int
	// BestLapCarModel is the car model used to achieve the fastest laptime
	BestLapCarModel int
	// CarData contains details about each car that this player drove on this track, keyed on car model
	CarData map[int]*PlayerTrackCarData
}

// Player describes a player which has driven in one or more events
type Player struct {
	// PlayerID is a unique ID for this player
	PlayerId string
	// MostRecentName contains the most recent name used by this player
	MostRecentName *Name
	// Names contains all names used by this player in any session
	Names []*Name
	// SessionNames contains the names of all sessions that this player took part in
	SessionNames []string
	// Events contains all events that this player took part in, keyed on event ID
	Events map[string]*Event
	// TrackData contains details about this player on each track they drove an event on, keyed on track label
	TrackData map[string]*PlayerTrackData
}

func newPlayerTrackCarData(player *Player, trackName string, carModel int) *PlayerTrackCarData {
	return &PlayerTrackCarData{
		player,
		trackName,
		carModel,
		math.MaxInt32,
	}
}

func newPlayerTrackData(player *Player, trackName string) *PlayerTrackData {
	return &PlayerTrackData{
		player,
		trackName,
		math.MaxInt32,
		0,
		make(map[int]*PlayerTrackCarData),
	}
}

func newPlayer(playerId string) *Player {
	return &Player{
		playerId,
		nil,
		nil,
		nil,
		make(map[string]*Event),
		make(map[string]*PlayerTrackData),
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

func (player *Player) addTrackDataForCarInSession(car *Car, session *Session) {
	track := accdata.TrackByLabel(session.TrackName)
	if track == nil {
		// log.Printf("Ignoring data for unknown track '%v'", session.TrackName)
		return
	}

	if player.TrackData[track.Label] == nil {
		player.TrackData[track.Label] = newPlayerTrackData(player, track.Label)
	}
	trackData := player.TrackData[track.Label]

	if trackData.CarData[car.CarModel] == nil {
		trackData.CarData[car.CarModel] = newPlayerTrackCarData(player, track.Label, car.CarModel)
	}
	carData := trackData.CarData[car.CarModel]

	for _, lap := range session.Laps {
		if (lap.CarId == car.CarId) && (car.Drivers[lap.DriverIndex].PlayerId == player.PlayerId) {
			if lap.Laptime < trackData.BestLap {
				trackData.BestLap = lap.Laptime
				trackData.BestLapCarModel = car.CarModel
			}

			if lap.Laptime < carData.BestLap {
				carData.BestLap = lap.Laptime
			}
		}
	}
}

func (player *Player) addCarInSession(car *Car, session *Session, event *Event) {
	player.SessionNames = append(player.SessionNames, session.SessionName)
	player.Events[event.EventId] = event
	player.addTrackDataForCarInSession(car, session)
}
