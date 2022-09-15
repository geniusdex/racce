package accserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/text/encoding/unicode"
)

// SessionType describes the type of a single session
type SessionType string

// DayOfWeekend represent a specific day in a race weekend
type DayOfWeekend int

// CarGroup defines the allowed car group in the server
type CarGroup string

const (
	Practice   SessionType = "P"
	Qualifying             = "Q"
	Race                   = "R"

	Friday   DayOfWeekend = 1
	Saturday              = 2
	Sunday                = 3

	FreeForAll CarGroup = "FreeForAll"
	GT3                 = "GT3"
	GT4                 = "GT4"
	GTC                 = "GTC"
	TCX                 = "TCX"
)

// CfgConfiguration contains the main server connectivity configuration.
type CfgConfiguration struct {
	UDPPort         int `json:"udpPort"`
	TCPPort         int `json:"tcpPort"`
	MaxConnections  int `json:"maxConnections"`
	LanDiscovery    int `json:"lanDiscovery"`
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
	PreRaceWaitingTimeSeconds int                `json:"preRaceWaitingTimeSeconds"`
	SessionOverTimeSeconds    int                `json:"sessionOverTimeSeconds"`
	AmbientTemp               int                `json:"ambientTemp"`
	CloudLevel                float32            `json:"cloudLevel"`
	Rain                      float32            `json:"rain"`
	WeatherRandomness         int                `json:"weatherRandomness"`
	PostQualySeconds          int                `json:"postQualySeconds"`
	PostRaceSeconds           int                `json:"postRaceSeconds"`
	Sessions                  []*CfgEventSession `json:"sessions"`
	MetaData                  string             `json:"metaData,omitempty"`
	ConfigVersion             int                `json:"configVersion"`
}

// CfgSettings contains generic server settings.
type CfgSettings struct {
	ServerName                 string   `json:"serverName"`
	AdminPassword              string   `json:"adminPassword"`
	CarGroup                   CarGroup `json:"carGroup"`
	TrackMedalsRequirement     int      `json:"trackMedalsRequirement"`
	SafetyRatingRequirement    int      `json:"safetyRatingRequirement"`
	RacecraftRatingRequirement int      `json:"racecraftRatingRequirement"`
	Password                   string   `json:"password"`
	SpectatorPassword          string   `json:"spectatorPassword"`
	MaxCarSlots                int      `json:"maxCarSlots"`
	DumpLeaderboards           int      `json:"dumpLeaderboards"`
	IsRaceLocked               int      `json:"isRaceLocked"`
	RandomizeTrackWhenEmpty    int      `json:"randomizeTrackWhenEmpty"`
	CentralEntryListPath       string   `json:"centralEntryListPath"`
	AllowAutoDQ                int      `json:"allowAutoDQ"`
	ShortFormationLap          int      `json:"shortFormationLap"`
	DumpEntryList              int      `json:"dumpEntryList"`
	FormationLapType           int      `json:"formationLapType"`
	IgnorePrematureDisconnects int      `json:"ignorePrematureDisconnects"`
	ConfigVersion              int      `json:"configVersion"`
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
	// Instance contains the running instance of the accServer, or nil if its not running
	Instance *Instance
	// LiveState contains the live state monitoring of the accServer; always available, even before first start
	LiveState *LiveState
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
		data, err = decoder.Bytes(data)
		if err != nil {
			return err
		}
	}

	return json.Unmarshal(data, target)
}

func writeCfgFile(path string, source interface{}) error {
	encoded, err := json.MarshalIndent(source, "", "\t")
	if err != nil {
		return err
	}

	encoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewEncoder()
	encoded, err = encoder.Bytes(encoded)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, encoded, 0644)
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
	if _, err := os.Stat(config.executable()); err != nil {
		return nil, fmt.Errorf("Could not locate accServer.exe: %v", err)
	}

	cfg, err := parseCfg(config.installationDir())
	if err != nil {
		return nil, fmt.Errorf("Cannot parse server config: %v", err)
	}

	return &Server{
		config,
		cfg,
		nil,
		newLiveState(),
	}, nil
}

// Start launches an instance of the server
func (s *Server) Start() error {
	if s.Instance.State() != Stopped {
		return fmt.Errorf("server is already running")
	}

	instance, err := newInstance(s.Config)
	if err != nil {
		return err
	}

	s.Instance = instance

	logParser := newLogParser(instance.NewLogChannel())
	s.LiveState.newInstance(logParser.Events)

	return nil
}

// Stop stops a running instance of the server
func (s *Server) Stop() error {
	return s.Instance.stop()
}

// IsRunning returns if the instance is running (State() == Running)
func (s *Server) IsRunning() bool {
	return s.Instance.State() == Running
}

// IsStopped returns if the instance is stopped (State() == Stopped)
func (s *Server) IsStopped() bool {
	return s.Instance.State() == Stopped
}

// IsStopping returns if the instance is stopping (State() == Stopping)
func (s *Server) IsStopping() bool {
	return s.Instance.State() == Stopping
}

// SaveConfiguration saves the current in-memory configuration to disk
func (s *Server) SaveConfiguration() error {
	cfgDir := s.Config.installationDir() + "/cfg/"

	if err := writeCfgFile(cfgDir+"configuration.json", s.Cfg.Configuration); err != nil {
		return fmt.Errorf("Cannot write cfg/configuration.json: %w", err)
	}

	if err := writeCfgFile(cfgDir+"settings.json", s.Cfg.Settings); err != nil {
		return fmt.Errorf("Cannot write cfg/settings.json: %w", err)
	}

	if err := writeCfgFile(cfgDir+"event.json", s.Cfg.Event); err != nil {
		return fmt.Errorf("Cannot write cfg/event.json: %w", err)
	}

	return nil
}
