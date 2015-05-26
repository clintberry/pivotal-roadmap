package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

var TOKEN = os.Getenv("TOKEN")
var PROJECT_ID = os.Getenv("PROJECT_ID")

type Label struct {
	ID        int    `json:"id"`
	ProjectID int    `json:"project_id"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Epic struct {
	ID          int    `json:"id"`
	Kind        string `json:"kind"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	ProjectID   int    `json:"project_id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Label       Label  `json:"label"`
	StartDate   time.Time
	ReleaseDate time.Time
	FinishDate  time.Time
}

type Story struct {
	Kind          string    `json:"kind"`
	ID            int       `json:"id"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
	AcceptedAt    string    `json:"accepted_at"`
	Estimate      int       `json:"estimate"`
	StoryType     string    `json:"story_type"`
	Name          string    `json:"name"`
	CurrentState  string    `json:"current_state"`
	RequestedByID int       `json:"requested_by_id"`
	ProjectID     int       `json:"project_id"`
	URL           string    `json:"url"`
	OwnerIds      []int     `json:"owner_ids"`
	Labels        []Label   `json:"labels"`
	OwnedByID     int       `json:"owned_by_id"`
	Deadline      time.Time `json:"deadline"`
}

type Iteration struct {
	Kind         string    `json:"kind"`
	Number       int       `json:"number"`
	ProjectID    int       `json:"project_id"`
	TeamStrength int       `json:"team_strength"`
	Stories      []Story   `json:"stories"`
	Start        time.Time `json:"start"`
	Finish       time.Time `json:"finish"`
}

func main() {
	epics, err := getEpics()
	//_, err := getEpics()
	if err != nil {
		fmt.Println("Couldn't Get Epics!!!")
	}

	iterations, err := getIterations(60)

	for _, iteration := range iterations {
		for _, story := range iteration.Stories {
			// fmt.Printf("%s : ", story.StoryType)
			// if story.StoryType == "release" {
			//  fmt.Printf("#%v", story)
			// }
			for _, label := range story.Labels {
				for eidx, epic := range epics {
					if label.Name == epic.Label.Name {
						switch story.StoryType {
						case "feature":
							if epic.StartDate.IsZero() {
								epics[eidx].StartDate = iteration.Start
							}
							epics[eidx].FinishDate = iteration.Finish
						case "release":
							epics[eidx].ReleaseDate = story.Deadline
						}
					}
				}
				//fmt.Printf("%s,", label.Name)
			}
			//fmt.Print("\n")
		}
	}

	for _, epic := range epics {
		fmt.Printf("\n%s - Start Date: %s, Release Date: %s, Finish Date: %s", epic.Name, epic.StartDate.Format("Jan 2, 2006"), epic.ReleaseDate.Format("Jan 2, 2006"), epic.FinishDate.Format("Jan 2, 2006"))
	}
}

func getEpics() ([]Epic, error) {
	var epics []Epic
	url := "https://www.pivotaltracker.com/services/v5/projects/" + PROJECT_ID + "/epics"
	// Build the request
	//
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-TrackerToken", TOKEN)
	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// Defer the closing of the body
	defer resp.Body.Close()
	// Read the content into a byte array
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &epics)
	if err != nil {
		return nil, err
	}
	// At this point we're done - simply return the bytes
	return epics, nil
}

func getIterations(offset int) ([]Iteration, error) {
	var iterations []Iteration
	url := "https://www.pivotaltracker.com/services/v5/projects/" + PROJECT_ID + "/iterations?offset=" + strconv.Itoa(offset)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-TrackerToken", TOKEN)
	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// Defer the closing of the body
	defer resp.Body.Close()
	// Read the content into a byte array
	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &iterations)
	if err != nil {
		return nil, err
	}
	// At this point we're done - simply return the bytes
	return iterations, nil
}

// func getEpicStartEndDates(label string) ([]Story, err) {

// }
