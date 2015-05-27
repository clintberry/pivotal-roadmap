package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Define a projectsIds type to get multiple project IDs from command line parameters
type projectids []string

func (p *projectids) String() string {
	return fmt.Sprint(*p)
}

func (p *projectids) Set(value string) error {
	for _, pid := range strings.Split(value, ",") {
		if strings.TrimSpace(pid) != "" {
			*p = append(*p, pid)
		}
	}
	return nil
}

// Implementation of new projects type
var projectsFlag projectids
var token string
var offset = 0

func init() {
	projectsFlag.Set(os.Getenv("PROJECTS"))
	flag.StringVar(&token, "token", "", "Pivotal tracker API token")
	flag.Var(&projectsFlag, "projects", "Comma-separated list of project IDs to add to roadmap")
	flag.IntVar(&offset, "offset", 0, "Start your iterations at a certain sprint number (ignore all before it)")
	if token == "" {
		token = os.Getenv("TOKEN")
	}
}

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

// Create epics type for sorting on the start date
type Epics []Epic

func (slice Epics) Len() int {
	return len(slice)
}

func (slice Epics) Less(i, j int) bool {
	// For epics with no start dates, put them at the end of the list
	// if slice[i].StartDate.IsZero() {
	// 	return false
	// }
	return slice[i].StartDate.Before(slice[j].StartDate)
}

func (slice Epics) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
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
	flag.Parse()
	fmt.Println(token)
	fmt.Println(projectsFlag)

	epics, err := getEpics()
	//_, err := getEpics()
	if err != nil {
		fmt.Println("Couldn't Get Epics!!!")
	}

	iterations, err := getIterations(offset)

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

	generateHtmlfile(Epics(epics))
}

func getEpics() ([]Epic, error) {
	var epics []Epic
	url := "https://www.pivotaltracker.com/services/v5/projects/" + projectsFlag[0] + "/epics"
	// Build the request
	//
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-TrackerToken", token)
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
	url := "https://www.pivotaltracker.com/services/v5/projects/" + projectsFlag[0] + "/iterations?offset=" + strconv.Itoa(offset)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-TrackerToken", token)
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

func generateHtmlfile(epics Epics) {
	sort.Sort(epics)
	fmt.Println("SORTED!!!!!!!!")
	for _, epic := range epics {
		fmt.Printf("\n%s - Start Date: %s, Release Date: %s, Finish Date: %s", epic.Name, epic.StartDate.Format("Jan 2, 2006"), epic.ReleaseDate.Format("Jan 2, 2006"), epic.FinishDate.Format("Jan 2, 2006"))
	}
}
