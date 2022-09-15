package accdata

// CompetitionSeries is a name for a series of competitions
type CompetitionSeries string

const (
	GTWorldChallenge            CompetitionSeries = "GT World Challenge"
	IntercontinentalGTChallenge                   = "Intercontinental GT Challenge"
	BritishGTChampionship                         = "British GT Championship"
	AmericanTrackPack                             = "American Track Pack"
)

// Competition describes a single competition
type Competition struct {
	// Series is the series to which this competition belongs
	Series CompetitionSeries
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
	// AlternateLabels are alternate labels used for this track, possibly used
	// in older server versions
	AlternateLabels []string
}

var (
	// Tracks contains information on all supported tracks
	Tracks = []*Track{
		{"monza", "Monza", Competition{GTWorldChallenge}, 29, 60,
			[]string{"monza_2019", "monza_2020"}},
		{"zolder", "Zolder", Competition{GTWorldChallenge}, 34, 50,
			[]string{"zolder_2019", "zolder_2020"}},
		{"brands_hatch", "Brands Hatch", Competition{GTWorldChallenge}, 32, 50,
			[]string{"brands_hatch_2019", "brands_hatch_2020"}},
		{"silverstone", "Silverstone", Competition{GTWorldChallenge}, 36, 60,
			[]string{"silverstone_2019", "silverstone_2020"}},
		{"paul_ricard", "Paul Ricard", Competition{GTWorldChallenge}, 33, 80,
			[]string{"paul_ricard_2019", "paul_ricard_2020"}},
		{"misano", "Misano", Competition{GTWorldChallenge}, 30, 50,
			[]string{"misano_2019", "misano_2020"}},
		{"spa", "Spa", Competition{GTWorldChallenge}, 82, 82,
			[]string{"spa_2019", "spa_2020"}},
		{"nurburgring", "Nurburgring", Competition{GTWorldChallenge}, 30, 50,
			[]string{"nurburgring_2019", "nurburgring_2020"}},
		{"barcelona", "Barcelona", Competition{GTWorldChallenge}, 29, 50,
			[]string{"barcelona_2019", "barcelona_2020"}},
		{"hungaroring", "Hungaroring", Competition{GTWorldChallenge}, 27, 50,
			[]string{"hungaroring_2019", "hungaroring_2020"}},
		{"zandvoort", "Zandvoort", Competition{GTWorldChallenge}, 25, 50,
			[]string{"zandvoort_2019", "zandvoort_2020"}},
		{"imola", "Imola", Competition{GTWorldChallenge}, 30, 50,
			[]string{"imola_2019", "imola_2020"}},
		{"kyalami", "Kyalami", Competition{IntercontinentalGTChallenge}, 40, 50,
			[]string{"kyalami_2019", "kyalami_2020"}},
		{"mount_panorama", "Mount Panorama", Competition{IntercontinentalGTChallenge}, 36, 50,
			[]string{"mount_panorama_2019", "mount_panorama_2020"}},
		{"suzuka", "Suzuka", Competition{IntercontinentalGTChallenge}, 51, 105,
			[]string{"suzuka_2019", "suzuka_2020"}},
		{"laguna_seca", "Laguna Seca", Competition{IntercontinentalGTChallenge}, 30, 50,
			[]string{"laguna_seca_2019", "laguna_seca_2020"}},
		{"oulton_park", "Oulton Park", Competition{BritishGTChampionship}, 28, 50,
			[]string{"oulton_park_2019", "oulton_park_2020"}},
		{"donington", "Donington Park", Competition{BritishGTChampionship}, 37, 50,
			[]string{"donington_2019", "donnington_2020"}},
		{"snetterton", "Snetterton", Competition{BritishGTChampionship}, 26, 50,
			[]string{"snetterton_2019", "snetterton_2020"}},
		{"cota", "Circuit of the Americas", Competition{AmericanTrackPack}, 30, 70,
			[]string{}},
		{"indianapolis", "Indianapolis", Competition{AmericanTrackPack}, 30, 60,
			[]string{}},
		{"watkins_glen", "Watkins Glen", Competition{AmericanTrackPack}, 30, 60,
			[]string{}},
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
			for _, label := range track.AlternateLabels {
				tracksByLabel[label] = track
			}
		}
	}
	return tracksByLabel[label]
}
