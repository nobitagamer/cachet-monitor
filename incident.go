package cachet

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
)

// Incident Cachet data model
type Incident struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
	Status  int    `json:"status"`
	Visible int    `json:"visible"`
	Notify  bool   `json:"notify"`

	ComponentID     int `json:"component_id"`
	ComponentStatus int `json:"component_status"`

	incidentTime time.Time
}

// IncidentResponse struct
type IncidentResponse struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Link        string      `json:"link"`
	Status      int64       `json:"status"`
	Order       int64       `json:"order"`
	GroupID     int64       `json:"group_id"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	DeletedAt   interface{} `json:"deleted_at"`
	Enabled     bool        `json:"enabled"`
	StatusName  string      `json:"status_name"`
}

// Send - Create or Update incident
func (incident *Incident) Send(cfg *CachetMonitor) (err error, updatedComponentStatus int) {

	// incidenet status is different server component status
	currentServerStatus, err := incident.GetComponentStatus(cfg)

	if err != nil {
		logrus.Warnf("cannot fetch component: %v", err)
	}

	// update status
	switch incident.Status {
	case 1, 2, 3:

		switch currentServerStatus {
		case 1:
			// change to major outage
			incident.ComponentStatus = 4
		case 2:
			// change to major outage
			incident.ComponentStatus = 4
		case 3:
			// change to major outage
			incident.ComponentStatus = 4
		case 4:
			// not change
			return nil, 0
		default:
			// not change
			return nil, 0
		}

	// incident fixed
	case 4:

		switch currentServerStatus {
		case 1:
			// not change
			return nil, 0
		case 2:
			// change to component alive
			incident.ComponentStatus = 1
		case 3:
			incident.ComponentStatus = 1
		case 4:
			incident.ComponentStatus = 1
		default:
			// not change
			return nil, 0
		}

	}

	requestType := "POST"
	requestURL := "/incidents"

	if incident.ID > 0 {
		requestType = "PUT"
		requestURL += "/" + strconv.Itoa(incident.ID)
	}

	jsonBytes, _ := json.Marshal(incident)

	resp, body, err := cfg.API.NewRequest(requestType, requestURL, jsonBytes)
	if err != nil {
		return err, 0
	}

	var respIncident = &IncidentResponse{}

	if err := json.Unmarshal(body.Data, respIncident); err != nil {
		return fmt.Errorf("Cannot parse incident body: %v, %v", err, string(body.Data)), 0
	}

	incident.ID = int(respIncident.ID)

	if resp.StatusCode != 200 {
		return fmt.Errorf("Could not create/update incident"), 0
	}

	return nil, incident.ComponentStatus
}

// GetComponentStatus func
func (incident *Incident) GetComponentStatus(cfg *CachetMonitor) (int, error) {
	resp, body, err := cfg.API.NewRequest("GET", "/components/"+strconv.Itoa(incident.ComponentID), nil)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("Invalid status code. Received %d, not found the component with id: %v", resp.StatusCode, incident.ComponentID)
	}

	data := IncidentResponse{}

	if err := json.Unmarshal(body.Data, &data); err != nil {
		return 0, fmt.Errorf("Cannot parse component body: %v. Err = %v", string(body.Data), err)
	}

	return int(data.Status), nil
}

// SetInvestigating sets status to Investigating
func (incident *Incident) SetInvestigating() {
	incident.Status = 1
}

// SetIdentified sets status to Identified
// not used now
func (incident *Incident) SetIdentified() {
	incident.Status = 2
}

// SetWatching sets status to Watching
// not used now
func (incident *Incident) SetWatching() {
	incident.Status = 3
}

// SetFixed sets status to Fixed
func (incident *Incident) SetFixed() {
	incident.Status = 4
}
