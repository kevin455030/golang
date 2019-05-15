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
)

type WeatherJson struct {
	//json檔結構
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
	}
	//天氣api
	apiUrl := "https://opendata.cwb.gov.tw/api/v1/rest/datastore/F-C0032-001"
	token := "your_token"
	elements := map[string]string{
		"locationName": r.FormValue("city"),
		"format":       "JSON",
	}
	res, err := http.Get(apiUrl + "?Authorization=" + token + "&locationName=" + elements["locationName"] + "&format=" + elements["format"])
	if err != nil {
	log.Fatal(err)
	}
	defer res.Body.Close()
	jsonData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	WeatherJson := WeatherJson{}
	json.Unmarshal([]byte(jsonData), &WeatherJson)

	wetherResult := WeatherJson.Records.Location[0].WeatherElement[0].Time[0].Parameter.ParameterName
	popResult := WeatherJson.Records.Location[0].WeatherElement[1].Time[0].Parameter.ParameterName
	minTResult := WeatherJson.Records.Location[0].WeatherElement[2].Time[0].Parameter.ParameterName
	maxTResult := WeatherJson.Records.Location[0].WeatherElement[4].Time[0].Parameter.ParameterName
	feelResult := WeatherJson.Records.Location[0].WeatherElement[3].Time[0].Parameter.ParameterName

	fmt.Fprintf(w, r.FormValue("city")+"\n")
	fmt.Fprintf(w, "天氣:"+wetherResult+"\n")
	fmt.Fprintf(w, "氣溫:"+minTResult+"°C~"+maxTResult+"°C"+"\n")
	fmt.Fprintf(w, "舒適度:"+feelResult+"\n")
	fmt.Fprintf(w, "降雨機率:"+popResult+"％"+"\n")

	//資料庫
	r.ParseForm()
	city := r.FormValue("city")
	db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/gomysql")
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

func main() {
	http.HandleFunc("/input", InputHandler)
	http.HandleFunc("/output", OutputHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
