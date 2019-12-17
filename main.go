package main

import "time"
import "log"
import "html/template"
import "net/http"
import "io/ioutil"
import "sort"
import "strconv"
import "balero_www/json2struct"

// trains is a slice of ATrain objects
var trains = []ATrain{}

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
		panic(err.Error())
	}
	return i
}

// The data on the page comes from this struct
type PageData struct {
	Trains []ATrain
}

// fetch the data from the api
func fetchTrains() []ATrain {
	var fetchedTrains = []ATrain{}
	url := "http://api.bart.gov/api/etd.aspx?cmd=etd&orig=mont&key=MW9S-E7SL-26DU-VV8V&dir=n&json=y"
	resp, err := http.Get(url)

	if err != nil {
		log.Panic(err)
	}

	trainData, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		log.Panic(err)
	}

	usableData := json2struct.RawDataIntoDataStruct(trainData)

	for _, train := range usableData.Root.Station[0].Etd {
		for _, est := range train.Est {
			minutes := convertStrToInt(est.Minutes)
			fetchedTrains = append(fetchedTrains, ATrain{train.Abbreviation, minutes})
		}
	}
	return fetchedTrains

}
func serveUI(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()
	trains = fetchTrains()
	trains = sortSlice(trains)
	tmpl, err := template.ParseFiles("templates/index.html")
	page := PageData{
		Trains: trains,
	}
	if err = tmpl.Execute(rw, page); err != nil {
		log.Panic("Failed to write template", err)
	}
	log.Printf("data served in %s\n", time.Since(start))
}
func sortSlice(train []ATrain) []ATrain {
	sort.Slice(train, func(i, j int) bool { return train[i].Minutes < train[j].Minutes })
	return train
}

func main() {
	address := ":8080"
	r := http.NewServeMux()
	r.HandleFunc("/", serveUI)
	log.Println(http.ListenAndServe(address, r))
}
