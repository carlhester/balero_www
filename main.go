package main

import "time"
import "log"
import "html/template"
import "net/http"
import "io/ioutil"
import "sort"
import "strconv"
import "balero_www/json2struct"

var trains = []ATrain{}

type ATrain struct {
	Train   string
	Minutes int
}

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

type PageData struct {
	Train []ATrain
}

func findATrain() {

	url := "http://api.bart.gov/api/etd.aspx?cmd=etd&orig=mont&key=MW9S-E7SL-26DU-VV8V&dir=n&json=y"
	resp, err := http.Get(url)

	if err != nil {
		log.Panic(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		log.Panic(err)
	}

	usableData := json2struct.RawDataIntoDataStruct(data)

	for _, train := range usableData.Root.Station[0].Etd {
		for _, est := range train.Est {
			minutes := convertStrToInt(est.Minutes)
			trains = append(trains, ATrain{train.Abbreviation, minutes})
			//fmt.Println(train.Abbreviation, minutes)
		}
	}
	trains = sortSlice(trains)

}
func serveUI(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()
	trains = trains[:0]
	findATrain()
	tmpl, err := template.ParseFiles("templates/index.html")
	page := PageData{
		Train: trains,
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
