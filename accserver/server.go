package accserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// SessionType describes the type of a single session
type SessionType string

// DayOfWeekend represent a specific day in a race weekend
type DayOfWeekend int

const (
	Practice   SessionType = "P"
	Qualifying             = "Q"
	Race                   = "R"

	Friday   DayOfWeekend = 1
	Saturday              = 2
	Sunday                = 3
)

// CfgConfiguration contains the main server connectivity configuration.
type CfgConfiguration struct {
	UDPPort         int `json:"udpPort"`
	TCPPort         int `json:"tcpPort"`
	MaxConnections  int `json:"maxConnections"`
	LanDiscover     int `json:"lanDiscover"`
	RegisterToLobby int `json:"registerToLobby"`
	ConfigVersion   int `json:"configVersion"`
}

// CfgEventSession contains the configuration for a single session within an event.
type CfgEventSession struct {
	HourOfDay              int         `json:"hourOfDay"`
	DayOfWeekend           int         `json:"dayOfWeekend"`
	TimeMultiplier         int         `json:"timeMultiplier"`
	SessionType            SessionType `json:"sessionType"`
	SessionDurationMinutes int         `json:"sessionDurationMinutes"`
}

// CfgEvent contains the configuration of the race event hosted by the server.
type CfgEvent struct {
	Track                     string             `json:"track"`
	PreRaceWiatingTimeSeconds int                `json:"preRaceWaitingTimeSeconds"`
	SessionOverTimeSeconds    int                `json:"sessionOverTimeSeconds"`
	AmbientTemp               int                `json:"ambientTemp"`
	CloudLevel                float32            `json:"cloudLevel"`
	Rain                      float32            `json:"rain"`
	WeatherRandomness         int                `json:"weatherRandomness"`
	Sessions                  []*CfgEventSession `json:"sessions"`
	MetaData                  string             `json:"metaData,omitempty"`
	ConfigVersion             int                `json:"configVersion"`
}

// CfgSettings contains generic server settings.
type CfgSettings struct {
	ServerName                 string `json:"serverName"`
	AdminPassword              string `json:"adminPassword"`
	TrackMedalsRequirement     int    `json:"trackMedalsRequirement"`
	SafetyRatingRequirement    int    `json:"safetyRatingRequirement"`
	RacecraftRatingRequirement int    `json:"racecraftRatingRequirement"`
	Password                   string `json:"password"`
	MaxCarSlots                int    `json:"maxCarSlots"`
	SpectatorPassword          string `json:"spectatorPassword"`
	DumpLeaderbords            int    `json:"dumpLeaderboards"`
	IsRaceLocked               int    `json:"isRaceLocked"`
	RandomizeTrackWhenEmpty    int    `json:"randomizeTrackWhenEmpty"`
	AllowAutoDQ                int    `json:"allowAutoDQ"`
	ShortFormationLap          int    `json:"shortFormationLap"`
	ConfigVersion              int    `json:"configVersion"`
}

// ServerConfiguration contains the parameters used by the accServer.
//
// It does not support all configuration files available yet.
type ServerConfiguration struct {
	Configuration *CfgConfiguration
	Settings      *CfgSettings
	Event         *CfgEvent
}

// Server represents an accServer installation, providing access to its
// configuration and allowing to start an instance.
type Server struct {
	// Config contains the meta-configuration on how the server is handled.
	Config *Configuration
	// Cfg contains the configuration used by the accServer instance itself.
	Cfg *ServerConfiguration
}

func isUtf16(data []byte) bool {
	return (data[0] == 0xFF && data[1] == 0xFE) || (data[0] == 0xFE && data[1] == 0xFF)
}

func parseCfgFile(path string, target interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if isUtf16(data) {
		decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
		data, _, err = transform.Bytes(decoder, data)
		if err != nil {
			return err
		}
	}

	return json.Unmarshal(data, target)
}

func parseCfg(installationPath string) (*ServerConfiguration, error) {
	cfg := &ServerConfiguration{
		&CfgConfiguration{},
		&CfgSettings{},
		&CfgEvent{},
	}

	if err := parseCfgFile(installationPath+"/cfg/configuration.json", cfg.Configuration); err != nil {
		return nil, fmt.Errorf("Cannot parse cfg/configuration.json: %v", err)
	}

	if err := parseCfgFile(installationPath+"/cfg/settings.json", cfg.Settings); err != nil {
		return nil, fmt.Errorf("Cannot parse cfg/settings.json: %v", err)
	}

	if err := parseCfgFile(installationPath+"/cfg/event.json", cfg.Event); err != nil {
		return nil, fmt.Errorf("Cannot parse cfg/event.json: %v", err)
	}

	return cfg, nil
}

// NewServer creates a new Server based on the given configuration.
func NewServer(config *Configuration) (*Server, error) {
	if _, err := os.Stat(config.installationDir() + "accServer.exe"); err != nil {
		return nil, fmt.Errorf("Could not locate accServer.exe: %v", err)
	}

	cfg, err := parseCfg(config.installationDir())
	if err != nil {
		return nil, fmt.Errorf("Cannot parse server config: %v", err)
	}

	return &Server{
		config,
		cfg,
	}, nil
}
