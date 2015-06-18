package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
)

type ProjectConfig struct {
	ProjectId string `toml:"project_id"`
	Offset    int    `toml:"offset"`
}
type Config struct {
	Token         string          `toml:"token"`
	ProjectConfig []ProjectConfig `toml:"project"`
}

var config Config
var token string

func init() {
	configBlob, err := ioutil.ReadFile("config.toml")
	fmt.Println(string(configBlob))
	if err != nil {
		log.Fatal(err)
	}
	if _, err := toml.Decode(string(configBlob), &config); err != nil {
		log.Fatal(err)
	}

	token = config.Token
}

type Project struct {
	AccountID                    int    `json:"account_id"`
	AtomEnabled                  bool   `json:"atom_enabled"`
	AutomaticPlanning            bool   `json:"automatic_planning"`
	BugsAndChoresAreEstimatable  bool   `json:"bugs_and_chores_are_estimatable"`
	CreatedAt                    string `json:"created_at"`
	CurrentIterationNumber       int    `json:"current_iteration_number"`
	Description                  string `json:"description"`
	EnableFollowing              bool   `json:"enable_following"`
	EnableIncomingEmails         bool   `json:"enable_incoming_emails"`
	EnableTasks                  bool   `json:"enable_tasks"`
	HasGoogleDomain              bool   `json:"has_google_domain"`
	ID                           int    `json:"id"`
	InitialVelocity              int    `json:"initial_velocity"`
	IterationLength              int    `json:"iteration_length"`
	Kind                         string `json:"kind"`
	Name                         string `json:"name"`
	NumberOfDoneIterationsToShow int    `json:"number_of_done_iterations_to_show"`
	PointScale                   string `json:"point_scale"`
	PointScaleIsCustom           bool   `json:"point_scale_is_custom"`
	ProfileContent               string `json:"profile_content"`
	ProjectType                  string `json:"project_type"`
	Public                       bool   `json:"public"`
	StartDate                    string `json:"start_date"`
	StartTime                    string `json:"start_time"`
	TimeZone                     struct {
		Kind      string `json:"kind"`
		Offset    string `json:"offset"`
		OlsonName string `json:"olson_name"`
	} `json:"time_zone"`
	UpdatedAt            string `json:"updated_at"`
	VelocityAveragedOver int    `json:"velocity_averaged_over"`
	Version              int    `json:"version"`
	WeekStartDay         string `json:"week_start_day"`
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
	ID                  int    `json:"id"`
	Kind                string `json:"kind"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
	ProjectID           int    `json:"project_id"`
	Name                string `json:"name"`
	URL                 string `json:"url"`
	Label               Label  `json:"label"`
	StartDate           time.Time
	ReleaseDate         time.Time
	FinishDate          time.Time
	StoryTotal          int
	StoryUnstartedTotal int
	StoryStartedTotal   int
	StoryFinishedTotal  int
	StoryDeliveredTotal int
	StoryAcceptedTotal  int
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
	TeamStrength float32   `json:"team_strength"`
	Stories      []Story   `json:"stories"`
	Start        time.Time `json:"start"`
	Finish       time.Time `json:"finish"`
}

func main() {

	var projectsHtml []string

	for _, project := range config.ProjectConfig {
		pid := project.ProjectId
		offset := project.Offset

		fmt.Printf("ProjectID: %s, Offset: %i", pid, offset)

		fmt.Println("PROJECT ID==========================" + pid)

		project, err := getProjectSettings(pid)
		if err != nil {
			fmt.Println("Couldn't load project settings")
		}

		epics, err := getEpics(pid)
		//_, err := getEpics()
		if err != nil {
			fmt.Println("Couldn't Get Epics!!!")
		}
		iterations, err := getIterations(pid, offset)

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

							epics[eidx].StoryTotal++

							switch story.CurrentState {
							case "unstarted":
								epics[eidx].StoryUnstartedTotal++
							case "started":
								epics[eidx].StoryStartedTotal++
							case "finished":
								epics[eidx].StoryFinishedTotal++
							case "delivered":
								epics[eidx].StoryDeliveredTotal++
							case "accepted":
								epics[eidx].StoryAcceptedTotal++
							}
						}
					}
					//fmt.Printf("%s,", label.Name)
				}
				//fmt.Print("\n")
			}
		}
		projectsHtml = append(projectsHtml, generateProjectHtml(project, Epics(epics), iterations))
	}
	generateHtmlfile(projectsHtml)
}

func getProjectSettings(projectId string) (Project, error) {
	var project Project
	url := "https://www.pivotaltracker.com/services/v5/projects/" + projectId

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return project, err
	}
	req.Header.Add("X-TrackerToken", token)
	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return project, err
	}
	// Defer the closing of the body
	defer resp.Body.Close()
	// Read the content into a byte array
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return project, err
	}
	err = json.Unmarshal(body, &project)
	if err != nil {
		return project, err
	}
	// At this point we're done - simply return the bytes
	return project, nil
}

func getEpics(projectId string) ([]Epic, error) {
	var epics []Epic
	url := "https://www.pivotaltracker.com/services/v5/projects/" + projectId + "/epics"
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

func getIterations(projectId string, offset int) ([]Iteration, error) {
	var iterations []Iteration
	// Have to hit the API multiple times for paginating until we get all results
	currentOffset := offset
	paginating := true
	for paginating == true {
		var paginatedIterations []Iteration
		url := "https://www.pivotaltracker.com/services/v5/projects/" + projectId + "/iterations?offset=" + strconv.Itoa(currentOffset) + "&limit=20"

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("X-TrackerToken", token)
		// Send the request via a client
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}

		paginationLimit, err := strconv.Atoi(resp.Header.Get("X-Tracker-Pagination-Limit"))
		paginationOffset, err := strconv.Atoi(resp.Header.Get("X-Tracker-Pagination-Offset"))
		paginationTotal, err := strconv.Atoi(resp.Header.Get("X-Tracker-Pagination-Total"))
		paginationReturned, err := strconv.Atoi(resp.Header.Get("X-Tracker-Pagination-Returned"))

		fmt.Println("Pagination Limit: " + strconv.Itoa(paginationLimit))
		fmt.Println("Pagination Offset: " + strconv.Itoa(paginationOffset))
		fmt.Println("Pagination Total: " + strconv.Itoa(paginationTotal))
		fmt.Println("Pagination Returned: " + strconv.Itoa(paginationReturned))
		fmt.Println(strconv.Itoa(paginationOffset + paginationReturned))

		// Defer the closing of the body
		defer resp.Body.Close()
		// Read the content into a byte array
		body, err := ioutil.ReadAll(resp.Body)
		//fmt.Println(string(body))
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}

		err = json.Unmarshal(body, &paginatedIterations)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}

		for _, iteration := range paginatedIterations {
			iterations = append(iterations, iteration)
		}
		// Have we gotten all results yet? If not, increment the offset and run again, otherwise break loop

		if (paginationOffset + paginationReturned) < paginationTotal {
			fmt.Println("More results available... let's go get 'em")
			currentOffset += paginationReturned
		} else {
			paginating = false
		}
	}

	// At this point we're done - simply return the bytes
	return iterations, nil
}

func generateProjectHtml(project Project, epics Epics, iterations []Iteration) string {
	sort.Sort(epics)

	for _, epic := range epics {
		fmt.Printf("\n%s - Start Date: %s, Release Date: %s, Finish Date: %s", epic.Name, epic.StartDate.Format("Jan 2, 2006"), epic.ReleaseDate.Format("Jan 2, 2006"), epic.FinishDate.Format("Jan 2, 2006"))
	}

	// baseCss, _ := ioutil.ReadFile("themes/boostrap.css")
	// themeCss, _ := ioutil.ReadFile("themes/bootstrap-theme.css")

	html := `<table class="table table-bordered roadmap">
	        <caption><h3>`

	html += project.Name

	html += `</h3></caption>
	            <thead><tr><th class="project-name">Feature</th>`

	for _, iteration := range iterations {
		html += "<th>" + iteration.Start.Format("Jan 2") + " - " + iteration.Finish.Format("Jan 2, 2006") + "</th>\n"
	}
	html += "</tr></thead><tbody>"

	for _, epic := range epics {
		var acceptedPercent float32
		var deliveredPercent float32
		var finishedPercent float32
		var startedPercent float32

		if !epic.StartDate.IsZero() {

			html += "<tr><td>" + epic.Name + "</td>"
			iterationStart := 0
			iterationFinish := len(iterations)
			iterationRelease := 0

			acceptedPercent = (float32(epic.StoryAcceptedTotal) / float32(epic.StoryTotal)) * 100
			deliveredPercent = (float32(epic.StoryDeliveredTotal) / float32(epic.StoryTotal)) * 100
			finishedPercent = (float32(epic.StoryFinishedTotal) / float32(epic.StoryTotal)) * 100
			startedPercent = (float32(epic.StoryStartedTotal) / float32(epic.StoryTotal)) * 100

			for index, iteration := range iterations {
				if iteration.Start == epic.StartDate {
					iterationStart = index
				}
				if iteration.Finish == epic.FinishDate {
					iterationFinish = index
				}
				if epic.ReleaseDate.After(iteration.Start) && epic.ReleaseDate.Before(iteration.Finish) {
					iterationRelease = index
				}
			}
			epicEnd := iterationFinish
			if iterationRelease != 0 && (iterationRelease > iterationFinish) {
				epicEnd = iterationRelease
			}

			for i := 0; i < len(iterations); i++ {

				if iterationStart == i {
					html += "<td colspan=\"" + strconv.Itoa(epicEnd-iterationStart+1) + "\">"
					html += "<div class='timeline'>&nbsp;"

					html += "<div class='timeline-accepted' style='width:" + strconv.Itoa(int(acceptedPercent)) + "%'>&nbsp</div>"
					html += "<div class='timeline-delivered' style='width:" + strconv.Itoa(int(deliveredPercent)) + "%'>&nbsp</div>"
					html += "<div class='timeline-finished' style='width:" + strconv.Itoa(int(finishedPercent)) + "%'>&nbsp</div>"
					html += "<div class='timeline-started' style='width:" + strconv.Itoa(int(startedPercent)) + "%'>&nbsp</div>"

					html += "</div>"
					i += (iterationFinish - iterationStart)
				} else {
					html += "<td>"
				}
				html += "</td>"

			}
			html += "</tr>"
		}
	}

	html += `
	        </tbody>
	        </table>`

	return html
}

func generateHtmlfile(projectHtml []string) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
<link rel="stylesheet" type="text/css" href="themes/bootstrap.css">
<link rel="stylesheet" type="text/css" href="themes/bootstrap-theme.css">

</head>
<body>
    <div style="overflow: scroll;">`
	for _, project := range projectHtml {
		html += project
	}
	html += `
    </div>
</body>
</html>`

	ioutil.WriteFile("roadmap.html", []byte(html), 0644)

	fmt.Print(html)
}
