package accdata

// CompetitionSeries is a name for a series of competitions over possibly multiple years
type CompetitionSeries string

const (
	GTWorldChallenge            CompetitionSeries = "GT World Challenge"
	IntercontinentalGTChallenge                   = "Intercontinental GT Challenge"
)

// Competition describes a single competition
type Competition struct {
	// Series is the series to which this competition belongs
	Series CompetitionSeries
	// Year is the year of the competition
	Year int
}

// Track represents a single track in the game
type Track struct {
	// Label is the label used for the track in server configuration or results
	Label string
	// Name is the properly formatted name for this track
	Name string
	// Competition indicates to which competition this track belongs
	Competition Competition
	// NrPitBoxes is the number of unique pit boxes on the track
	NrPitBoxes int
	// PrivateServerSlots indicates how many car slots are available for this
	// track on private servers
	PrivateServerSlots int
}

var (
	// Tracks contains information on all supported tracks
	Tracks = []*Track{
		&Track{"monza", "Monza", Competition{GTWorldChallenge, 2018}, 29, 60},
		&Track{"zolder", "Zolder", Competition{GTWorldChallenge, 2018}, 34, 50},
		&Track{"brands_hatch", "Brands Hatch", Competition{GTWorldChallenge, 2018}, 32, 50},
		&Track{"silverstone", "Silverstone", Competition{GTWorldChallenge, 2018}, 36, 50},
		&Track{"paul_ricard", "Paul Ricard", Competition{GTWorldChallenge, 2018}, 33, 60},
		&Track{"misano", "Misano", Competition{GTWorldChallenge, 2018}, 30, 50},
		&Track{"spa", "Spa", Competition{GTWorldChallenge, 2018}, 82, 82},
		&Track{"nurburgring", "Nurburgring", Competition{GTWorldChallenge, 2018}, 30, 50},
		&Track{"barcelona", "Barcelona", Competition{GTWorldChallenge, 2018}, 29, 50},
		&Track{"hungaroring", "Hungaroring", Competition{GTWorldChallenge, 2018}, 27, 50},
		&Track{"zandvoort", "Zandvoort", Competition{GTWorldChallenge, 2018}, 25, 50},
		&Track{"monza_2019", "Monza", Competition{GTWorldChallenge, 2019}, 29, 60},
		&Track{"zolder_2019", "Zolder", Competition{GTWorldChallenge, 2019}, 34, 50},
		&Track{"brands_hatch_2019", "Brands Hatch", Competition{GTWorldChallenge, 2019}, 32, 50},
		&Track{"silverstone_2019", "Silverstone", Competition{GTWorldChallenge, 2019}, 36, 50},
		&Track{"paul_ricard_2019", "Paul Ricard", Competition{GTWorldChallenge, 2019}, 33, 60},
		&Track{"misano_2019", "Misano", Competition{GTWorldChallenge, 2019}, 30, 50},
		&Track{"spa_2019", "Spa", Competition{GTWorldChallenge, 2019}, 82, 82},
		&Track{"nurburgring_2019", "Nurburgring", Competition{GTWorldChallenge, 2019}, 30, 50},
		&Track{"barcelona_2019", "Barcelona", Competition{GTWorldChallenge, 2019}, 29, 50},
		&Track{"hungaroring_2019", "Hungaroring", Competition{GTWorldChallenge, 2019}, 27, 50},
		&Track{"zandvoort_2019", "Zandvoort", Competition{GTWorldChallenge, 2019}, 25, 50},
		&Track{"kyalami_2019", "Kyalami", Competition{IntercontinentalGTChallenge, 2019}, 40, 50},
		&Track{"mount_panorama_2019", "Mount Panorama", Competition{IntercontinentalGTChallenge, 2019}, 36, 50},
		&Track{"suzuka_2019", "Suzuka", Competition{IntercontinentalGTChallenge, 2019}, 51, 105},
		&Track{"laguna_seca_2019", "Laguna Seca", Competition{IntercontinentalGTChallenge, 2019}, 30, 50},
		&Track{"monza_2020", "Monza", Competition{GTWorldChallenge, 2020}, 29, 60},
		&Track{"zolder_2020", "Zolder", Competition{GTWorldChallenge, 2020}, 34, 50},
		&Track{"brands_hatch_2020", "Brands Hatch", Competition{GTWorldChallenge, 2020}, 32, 50},
		&Track{"silverstone_2020", "Silverstone", Competition{GTWorldChallenge, 2020}, 36, 50},
		&Track{"paul_ricard_2020", "Paul Ricard", Competition{GTWorldChallenge, 2020}, 33, 60},
		&Track{"misano_2020", "Misano", Competition{GTWorldChallenge, 2020}, 30, 50},
		&Track{"spa_2020", "Spa", Competition{GTWorldChallenge, 2020}, 82, 82},
		&Track{"nurburgring_2020", "Nurburgring", Competition{GTWorldChallenge, 2020}, 30, 50},
		&Track{"barcelona_2020", "Barcelona", Competition{GTWorldChallenge, 2020}, 29, 50},
		&Track{"hungaroring_2020", "Hungaroring", Competition{GTWorldChallenge, 2020}, 27, 50},
		&Track{"zandvoort_2020", "Zandvoort", Competition{GTWorldChallenge, 2020}, 25, 50},
		&Track{"imola_2020", "Imola", Competition{GTWorldChallenge, 2020}, 30, 50},
	}

	// tracksByLabel is a cache for TrackByLabel
	tracksByLabel map[string]*Track
)

// TrackByLabel returns the track for a given label
func TrackByLabel(label string) *Track {
	if tracksByLabel == nil {
		tracksByLabel = make(map[string]*Track)
		for _, track := range Tracks {
			tracksByLabel[track.Label] = track
		}
	}
	return tracksByLabel[label]
}
