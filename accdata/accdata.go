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
	Tracks = map[string]*Track{
		"monza":               &Track{"Monza", Competition{GTWorldChallenge, 2018}, 29, 60},
		"zolder":              &Track{"Zolder", Competition{GTWorldChallenge, 2018}, 34, 50},
		"brands_hatch":        &Track{"Brands Hatch", Competition{GTWorldChallenge, 2018}, 32, 50},
		"silverstone":         &Track{"Silverstone", Competition{GTWorldChallenge, 2018}, 36, 50},
		"paul_ricard":         &Track{"Paul Ricard", Competition{GTWorldChallenge, 2018}, 33, 60},
		"misano":              &Track{"Misano", Competition{GTWorldChallenge, 2018}, 30, 50},
		"spa":                 &Track{"Spa-Francorchamps", Competition{GTWorldChallenge, 2018}, 82, 82},
		"nurburgring":         &Track{"Nurburgring", Competition{GTWorldChallenge, 2018}, 30, 50},
		"barcelona":           &Track{"Barcelona", Competition{GTWorldChallenge, 2018}, 29, 50},
		"hungaroring":         &Track{"Hungaroring", Competition{GTWorldChallenge, 2018}, 27, 50},
		"zandvoort":           &Track{"Zandvoort", Competition{GTWorldChallenge, 2018}, 25, 50},
		"monza_2019":          &Track{"Monza", Competition{GTWorldChallenge, 2019}, 29, 60},
		"zolder_2019":         &Track{"Zolder", Competition{GTWorldChallenge, 2019}, 34, 50},
		"brands_hatch_2019":   &Track{"Brands Hatch", Competition{GTWorldChallenge, 2019}, 32, 50},
		"silverstone_2019":    &Track{"Silverstone", Competition{GTWorldChallenge, 2019}, 36, 50},
		"paul_ricard_2019":    &Track{"Paul Ricard", Competition{GTWorldChallenge, 2019}, 33, 60},
		"misano_2019":         &Track{"Misano", Competition{GTWorldChallenge, 2019}, 30, 50},
		"spa_2019":            &Track{"Spa-Francorchamps", Competition{GTWorldChallenge, 2019}, 82, 82},
		"nuburgring_2019":     &Track{"Nurburgring", Competition{GTWorldChallenge, 2019}, 30, 50},
		"barcelona_2019":      &Track{"Barcelona", Competition{GTWorldChallenge, 2019}, 29, 50},
		"hungaroring_2019":    &Track{"Hungaroring", Competition{GTWorldChallenge, 2019}, 27, 50},
		"zandvoort_2019":      &Track{"Zandvoort", Competition{GTWorldChallenge, 2019}, 25, 50},
		"kyalami_2019":        &Track{"Kyalami", Competition{IntercontinentalGTChallenge, 2019}, 40, 50},
		"mount_panorama_2019": &Track{"Mount Panorama", Competition{IntercontinentalGTChallenge, 2019}, 36, 50},
		"suzuka_2019":         &Track{"Suzuka", Competition{IntercontinentalGTChallenge, 2019}, 51, 105},
		"laguna_seca_2019":    &Track{"Laguna Seca", Competition{IntercontinentalGTChallenge, 2019}, 30, 50},
	}
)
