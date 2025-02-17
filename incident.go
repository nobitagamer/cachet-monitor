package cachet

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
)

// Incident Cachet data model
type Incident struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
	Status  int    `json:"status"`
	Visible int    `json"visible"`
	Notify  bool   `json:"notify"`

	ComponentID     int `json:"component_id"`
	ComponentStatus int `json:"component_status"`
}

//Get the last still open incident
func (mon *AbstractMonitor) Get(cfg *CachetMonitor) (*Incident, error) {

	requestType := "GET"
	requestURL := fmt.Sprintf("/incidents?component_id=%d", mon.ComponentID)
	_, body, err := cfg.API.NewRequest(requestType, requestURL, nil)
	if err != nil {
		return nil, err
	}
	data := make([]Incident, 0)
	if err := json.Unmarshal(body.Data, &data); err != nil {
		return nil, fmt.Errorf("Cannot parse incident body: %v, %v", err, string(body.Data))
	}
	//filter out resolved incidents
	openIncidents := make([]Incident, 0)
	for _, i := range data {
		if i.Status < 4 {
			openIncidents = append(openIncidents, i)
		}
	}
	if len(openIncidents) == 0 {
		return nil, nil
	}
	return &openIncidents[0], nil

}

// Send - Create or Update incident
func (incident *Incident) Send(cfg *CachetMonitor) error {
	switch incident.Status {
	case 1, 2, 3:
		// partial outage
		incident.ComponentStatus = 3

		componentStatus, err := incident.GetComponentStatus(cfg)
		if componentStatus == 3 {
			// major outage
			incident.ComponentStatus = 4
		}

		if err != nil {
			logrus.Warnf("cannot fetch component: %v", err)
		}
	case 4:
		// fixed
		incident.ComponentStatus = 1
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
		return err
	}

	var data struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(body.Data, &data); err != nil {
		return fmt.Errorf("Cannot parse incident body: %v, %v", err, string(body.Data))
	}

	incident.ID = data.ID
	if resp.StatusCode != 200 {
		return fmt.Errorf("Could not create/update incident!")
	}

	return nil
}

func (incident *Incident) GetComponentStatus(cfg *CachetMonitor) (int, error) {
	resp, body, err := cfg.API.NewRequest("GET", "/components/"+strconv.Itoa(incident.ComponentID), nil)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("Invalid status code. Received %d", resp.StatusCode)
	}

	var data struct {
		Status int `json:"status,string"`
	}
	if err := json.Unmarshal(body.Data, &data); err != nil {
		return 0, fmt.Errorf("Cannot parse component body: %v. Err = %v", string(body.Data), err)
	}

	return data.Status, nil
}

// SetInvestigating sets status to Investigating
func (incident *Incident) SetInvestigating() {
	incident.Status = 1
}

// SetIdentified sets status to Identified
func (incident *Incident) SetIdentified() {
	incident.Status = 2
}

// SetWatching sets status to Watching
func (incident *Incident) SetWatching() {
	incident.Status = 3
}

// SetFixed sets status to Fixed
func (incident *Incident) SetFixed() {
	incident.Status = 4
}
