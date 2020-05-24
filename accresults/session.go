package accresults

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type SessionType string

const (
	Practice   SessionType = "P"
	Qualifying             = "Q"
	Race                   = "R"
)

// TODO: enum CarModel
// TODO: enum CupCategory
// TODO: enum Nationality

type Driver struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	ShortName string `json:"shortName"`
	PlayerId  string `json:"playerId"`
}

type Car struct {
	CarId       int       `json:"carId"`
	RaceNumber  int       `json:"raceNumber"`
	CarModel    int       `json:"carModel"`
	CupCategory int       `json:"cupCategory"`
	TeamName    string    `json:"teamName"`
	Nationality int       `json:"nationality"`
	CarGuid     int       `json:"carGuid"`
	TeamGuid    int       `json:"teamGuid"`
	Drivers     []*Driver `json:"drivers"`
}

type LeaderBoardTiming struct {
	// TODO: durations
	LastLap     int    `json:"lastLap"`
	LastSplits  []*int `json:"lastSplits"`
	BestLap     int    `json:"bestLap"`
	BestSplits  []*int `json:"bestSplits"`
	TotalTime   int    `json:"totalTime"`
	LapCount    int    `json:"lapCount"`
	LastSplitId int    `json:"lastSplitId"`
}

type LeaderBoardLine struct {
	Car                     *Car               `json:"car"`
	CurrentDriver           *Driver            `json:"currentDriver"`
	CurrentDriverIndex      int                `json:"currentDriverIndex"`
	Timing                  *LeaderBoardTiming `json:"timing"`
	MissingMandatoryPitstop int                `json:"missingMandatoryPitstop`
	//    DriverTotalTimes []*time.Duration `json:"driverTotalTimes"`
	DriverTotalTimes []*float64 `json:"driverTotalTimes"`
}

type SessionResult struct {
	//    BestLap *time.Duration `json:"bestLap"`
	BestLap int `json:"bestLap"`
	//    BestSplits []*time.Duration `json:"bestSplits"`
	BestSplits       []*int             `json:"bestSplits"`
	IsWetSession     int                `json:"isWetSession"`
	Type             int                `json:"type"`
	LeaderBoardLines []*LeaderBoardLine `json:"leaderBoardLines"`

	// Indexed on CarId, the 1-based position after every lap it completed
	CarPositionsPerLap map[int][]int
}

type Lap struct {
	CarId       int `json:"carId"`
	DriverIndex int `json:"driverIndex"`
	//    Laptime time.Duration `json:"laptime"`
	Laptime        int  `json:"laptime"`
	IsValidForBest bool `json:"isValidForBest"`
	//    Splits []time.Duration `json:"splits"`
	Splits []int `json:"splits"`
}

type Penalty struct {
	CarId          int    `json:"carId"`
	DriverIndex    int    `json:"driverIndex"`
	Reason         string `json:"reason"`
	Penalty        string `json:"penalty"`
	PenaltyValue   int    `json:"penaltyValue"`
	ViolationInLap int    `json:"violationInLap"`
	ClearedInLap   int    `json:"clearedInLap"`
}

type Session struct {
	TrackName         string         `json:"trackName"`
	SessionType       SessionType    `json:"sessionType"`
	SessionIndex      int            `json:"sessionIndex"`
	RaceWeekendIndex  int            `json:"raceWeekendIndex"`
	MetaData          string         `json:"metaData"`
	ServerName        string         `json:"serverName"`
	SessionResult     *SessionResult `json:"sessionResult"`
	Laps              []*Lap         `json:"laps"`
	Penalties         []*Penalty     `json:"penalties"`
	PostRacePenalties []*Penalty     `json:"post_race_penalties"`

	SessionName       string
	EndTime           time.Time
	SessionTypeString string
}

func (session *Session) calculateCarPositionsPerLap() {
	carPositionsPerLap := make(map[int][]int)
	carsPerLap := make(map[int]int)
	for _, lap := range session.Laps {
		lapNr := len(carPositionsPerLap[lap.CarId])
		carsPerLap[lapNr]++
		pos := carsPerLap[lapNr]
		carPositionsPerLap[lap.CarId] = append(carPositionsPerLap[lap.CarId], pos)
	}
	session.SessionResult.CarPositionsPerLap = carPositionsPerLap
}

func (session *Session) FindCarById(carId int) *Car {
	for _, line := range session.SessionResult.LeaderBoardLines {
		if line.Car.CarId == carId {
			return line.Car
		}
	}
	return nil
}

func readUtf16File(filename string) ([]byte, error) {
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	decodedContents, _, err2 := transform.Bytes(decoder, fileContents)
	return decodedContents, err2
}

func LoadSessionFromFile(filename string, endTime time.Time) (*Session, error) {
	fileContents, err := readUtf16File(filename)
	if err != nil {
		return nil, err
	}

	var session Session
	err = json.Unmarshal(fileContents, &session)
	if err != nil {
		return nil, err
	}

	session.EndTime = endTime
	session.calculateCarPositionsPerLap()

	return &session, nil
}
