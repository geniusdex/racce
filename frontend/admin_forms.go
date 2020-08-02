package frontend

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/geniusdex/racce/accserver"
)

type errorStore struct {
	errors []string
}

func newErrorStore() *errorStore {
	return &errorStore{make([]string, 0)}
}

func (s *errorStore) Add(err error) {
	if err != nil {
		s.errors = append(s.errors, err.Error())
	}
}

func (s *errorStore) Error() error {
	if len(s.errors) == 0 {
		return nil
	}

	return fmt.Errorf(strings.Join(s.errors, "\n"))
}

type formParser struct {
	form   url.Values
	errors *errorStore
}

func newFormParser(form url.Values) *formParser {
	return &formParser{form, newErrorStore()}
}

func (p *formParser) Error() error {
	return p.errors.Error()
}

// get looks up a form value and returns the value plus whether it was found.
// If the form value was not found, the value is stored already via addError()
func (p *formParser) get(field string) (string, bool) {
	if values, ok := p.form[field]; !ok || len(values) == 0 {
		p.errors.Add(fmt.Errorf("Missing field %s", field))
		return "", false
	} else {
		return values[0], true
	}
}

func (p *formParser) String(field string) string {
	strValue, _ := p.get(field)
	return strValue
}

func (p *formParser) Int(field string) int {
	if strValue, ok := p.get(field); !ok {
		return 0
	} else {
		intValue, err := strconv.Atoi(strValue)
		if err != nil {
			p.errors.Add(fmt.Errorf("Field %s is not a valid integer: %w", field, err))
		}
		return intValue
	}
}

func (p *formParser) Float32(field string) float32 {
	if strValue, ok := p.get(field); !ok {
		return 0
	} else {
		floatValue, err := strconv.ParseFloat(strValue, 32)
		if err != nil {
			p.errors.Add(fmt.Errorf("Field %s is not a valid number: %w", field, err))
		}
		return float32(floatValue)
	}
}

// parseServerCfgEventFormSessions parses the sessions in the given event configuration form
func (a *admin) parseServerCfgEventFormSessions(form url.Values) ([]*accserver.CfgEventSession, error) {
	// Figure out all session indices
	sessionIDs := make(map[string]bool)
	for field := range form {
		// suffix is not ok
		if strings.HasPrefix(field, "sessions[") {
			sessionID := strings.TrimPrefix(field[:strings.Index(field, "]")], "sessions[")
			sessionIDs[sessionID] = true
		}
	}

	// Sort them on numeric value for proper session ordering
	sortedSessionIDs := make([]string, 0)
	for sessionID := range sessionIDs {
		sortedSessionIDs = append(sortedSessionIDs, sessionID)
		sort.Slice(sortedSessionIDs, func(i, j int) bool {
			a, _ := strconv.Atoi(sortedSessionIDs[i])
			b, _ := strconv.Atoi(sortedSessionIDs[j])
			return a < b
		})
	}

	// There must be at least 1 sessions
	var sessions = make([]*accserver.CfgEventSession, len(sortedSessionIDs))
	if len(sortedSessionIDs) == 0 {
		return sessions, fmt.Errorf("No sessions defined")
	}

	// Create and parse sessions
	parser := newFormParser(form)
	for i, sessionID := range sortedSessionIDs {
		prefix := "sessions[" + sessionID + "]."
		sessions[i] = &accserver.CfgEventSession{
			HourOfDay:              parser.Int(prefix + "hourOfDay"),
			DayOfWeekend:           parser.Int(prefix + "dayOfWeekend"),
			TimeMultiplier:         parser.Int(prefix + "timeMultiplier"),
			SessionType:            accserver.SessionType(parser.String(prefix + "sessionType")),
			SessionDurationMinutes: parser.Int(prefix + "sessionDurationMinutes"),
		}
	}

	return sessions, parser.Error()
}

// parseServerCfgEventForm parses the given event configuration form and returns a new event configuration
func (a *admin) parseServerCfgEventForm(form url.Values) (*accserver.CfgEvent, error) {
	parser := newFormParser(form)

	event := &accserver.CfgEvent{
		Track:                     parser.String("track"),
		PreRaceWaitingTimeSeconds: parser.Int("preRaceWaitingTimeSeconds"),
		SessionOverTimeSeconds:    parser.Int("sessionOverTimeSeconds"),
		AmbientTemp:               parser.Int("ambientTemp"),
		CloudLevel:                parser.Float32("cloudLevel"),
		Rain:                      parser.Float32("rain"),
		WeatherRandomness:         parser.Int("weatherRandomness"),
		// event.PostQualySeconds = parser.Int("postQualySeconds")
		// event.PostRaceSeconds = parser.Int("PostRaceSeconds")
		// event.MetaData = ??
		ConfigVersion: 1,
	}

	errors := newErrorStore()
	errors.Add(parser.Error())
	var err error
	event.Sessions, err = a.parseServerCfgEventFormSessions(form)
	errors.Add(err)

	return event, errors.Error()
}
