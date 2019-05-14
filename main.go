package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type WetherJson struct {
	Success string `json:"success"`
	Result  struct {
		ResourceID string `json:"resource_id"`
		Fields     []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"fields"`
	} `json:"result"`
	Records struct {
		DatasetDescription string `json:"datasetDescription"`
		Location           []struct {
			LocationName   string `json:"locationName"`
			WeatherElement []struct {
				ElementName string `json:"elementName"`
				Time        []struct {
					StartTime string `json:"startTime"`
					EndTime   string `json:"endTime"`
					Parameter struct {
						ParameterName  string `json:"parameterName"`
						ParameterValue string `json:"parameterValue"`
					} `json:"parameter"`
				} `json:"time"`
			} `json:"weatherElement"`
		} `json:"location"`
	} `json:"records"`
}

func InputHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t, err := template.ParseFiles("input.html")
		if err != nil {
			log.Println(err)
		}
		err = t.Execute(w, nil)
		if err != nil {
			log.Println(err)
		}
	}
}

func OutputHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		t, err := template.ParseFiles("output.html")
		if err != nil {
			log.Println(err)
		}
		err = t.Execute(w, nil)
		if err != nil {
			log.Println(err)
		}
		apiUrl := "https://opendata.cwb.gov.tw/api/v1/rest/datastore/F-C0032-001"
		token := "CWB-95C28952-5740-4990-A976-0E14F972C8F2"
		elements := map[string]string{
			"locationName": r.FormValue("city"),
			"format":       "JSON",
		}
		res, err := http.Get(apiUrl + "?Authorization=" + token + "&locationName=" + elements["locationName"] + "&format=" + elements["format"])
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		sitemap, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("%s", sitemap)
		WetherJson := WetherJson{}
		json.Unmarshal([]byte(sitemap), &WetherJson)

		//Time[Wx,PoP,MinT,Ci,MaxT]
		wetherResult := WetherJson.Records.Location[0].WeatherElement[0].Time[0].Parameter.ParameterName
		popResult := WetherJson.Records.Location[0].WeatherElement[1].Time[0].Parameter.ParameterName
		minTResult := WetherJson.Records.Location[0].WeatherElement[2].Time[0].Parameter.ParameterName
		maxTResult := WetherJson.Records.Location[0].WeatherElement[4].Time[0].Parameter.ParameterName
		feelResult := WetherJson.Records.Location[0].WeatherElement[3].Time[0].Parameter.ParameterName

		fmt.Fprintf(w, r.FormValue("city")+"\n")
		fmt.Fprintf(w, "天氣:"+wetherResult+"\n")
		fmt.Fprintf(w, "氣溫:"+minTResult+"°C~"+maxTResult+"°C"+"\n")
		fmt.Fprintf(w, "舒適度:"+feelResult+"\n")
		fmt.Fprintf(w, "降雨機率:"+popResult+"％"+"\n")

		r.ParseForm()
		city := r.FormValue("city")
		db, err := sql.Open("mysql", "root:abc123@tcp(localhost:3306)/gomysql")
		if err != nil {
			panic(err.Error())
		}
		defer db.Close()
		insert, err := db.Query("INSERT INTO city VALUES('" + city + "','" + wetherResult + "','" + popResult + "','" + minTResult + "','" + maxTResult + "','" + feelResult + "')")
		if err != nil {
			panic(err.Error())
		}
		defer insert.Close()
		fmt.Println("新增資料庫成功!!")
	}

}

func main() {
	http.HandleFunc("/input", InputHandler)
	http.HandleFunc("/output", OutputHandler)
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
