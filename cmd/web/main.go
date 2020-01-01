package main

import "time"
import "log"
import "html/template"
import "net/http"
import "io/ioutil"
import "sort"
import "strconv"

// trains is a slice of ATrain objects
var trains = []ATrain{}
var stations = []AStation{}

var allStations = []string{"MONT", "WCRK", "ANTC"}
var allDirections = []string{"n", "s"}
var allLines = []string{"yellow", "red", "blue"}

var selectedStation string = "mont"
var selectedDirection string
var selectedLine string

// ATrain is a single train
type ATrain struct {
	Train   string
	Minutes int
}

// Minutes is returned from api as a string, convert to int
func convertStrToInt(input string) int {
	if input == "Leaving" {
		input = "0"
	}

	i, err := strconv.Atoi(input)
	if err != nil {
		log.Panic(err.Error())
	}
	return i
}

// The data on the page comes from this struct
type PageData struct {
	Trains   []ATrain
	Stations []AStation
}

type AStation struct {
	Name string
}

// fetch the data from the api
func fetchTrains(station string) []ATrain {
	start := time.Now()
	log.Printf("[fetchTrains] START. Passed %s", station)

	var fetchedTrains = []ATrain{}
	url := "http://api.bart.gov/api/etd.aspx?cmd=etd&orig=" + station + "&key=MW9S-E7SL-26DU-VV8V&dir=n&json=y"
	log.Printf("[fetchTrains] URL: %s", url)
	resp, err := http.Get(url)

	if err != nil {
		log.Panic(err)
	}

	trainData, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		log.Panic(err)
	}

	usableData := RawDataIntoDataStruct(trainData)

	for _, train := range usableData.Root.Station[0].Etd {
		for _, est := range train.Est {
			minutes := convertStrToInt(est.Minutes)
			fetchedTrains = append(fetchedTrains, ATrain{train.Abbreviation, minutes})
		}
	}
	log.Printf("[fetchTrains] returning %d entries", len(fetchedTrains))
	log.Printf("[fetchTrains] END: %s\n\n", time.Since(start))
	return fetchedTrains
}

func fetchStations() []AStation {
	start := time.Now()
	log.Printf("[fetchStations] START")
	var fetchedStations []AStation
	var fetchedStationsStr []string

	for _, station := range allStations {
		fetchedStations = append(fetchedStations, AStation{station})
	}

	for _, stat := range fetchedStations {
		fetchedStationsStr = append(fetchedStationsStr, stat.Name)
	}

	log.Printf("[fetchStations] returning %d entries (strings)	: %s", len(fetchedStationsStr), fetchedStationsStr)
	log.Printf("[fetchStations] returning %d entries			: %s\n", len(fetchedStations), fetchedStations)
	log.Printf("[fetchStations] END: %s\n\n", time.Since(start))
	return fetchedStations
}

func fetchSelectedStation(r *http.Request) string {
	start := time.Now()
	log.Printf("[fetchSelectedStation] START")
	r.ParseForm()
	//station := fmt.Sprintf("%s", r.Form["station"])
	station := r.Form["station"][0]
	log.Printf("[fetchSelectedStation] returning: %s", station)
	log.Printf("[fetchSelectedStation] END: %s\n\n", time.Since(start))
	return station
}

func updateUI(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Printf("[updateUI] START")
	log.Printf("[updateUI] %+v", r)
	rw.Header().Set("Cache-Control", "no-store")
	selectedStation = fetchSelectedStation(r)
	http.Redirect(rw, r, "http://"+r.Host, 301) // Send browser back to main page
	log.Printf("[updateUI] END: %s\n\n", time.Since(start))
}

func serveUI(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Printf("[serveUI] START")
	log.Printf("[serveUI] %+v", r)
	log.Printf("[serveUI] selected station : %s", selectedStation)
	rw.Header().Set("Cache-Control", "no-store")

	stations = fetchStations()
	trains = fetchTrains(selectedStation)
	trains = sortSlice(trains)

	tmpl, err := template.ParseFiles("ui/html/index.html")
	page := PageData{
		Trains:   (trains),
		Stations: stations,
	}
	if err = tmpl.Execute(rw, page); err != nil {
		log.Panic("Failed to write template", err)
	}
	log.Printf("[serveUI] %+v", rw)
	log.Printf("[serveUI] END: %s\n\n", time.Since(start))
}
func sortSlice(train []ATrain) []ATrain {
	sort.Slice(train, func(i, j int) bool { return train[i].Minutes < train[j].Minutes })
	return train
}

func main() {
	address := ":8080"
	r := http.NewServeMux()
	r.HandleFunc("/update", updateUI)
	r.HandleFunc("/", serveUI)
	log.Println(http.ListenAndServe(address, r))
}
