package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
)

type VostPlaylist struct {
	Std     string `json:"std"`
	Preview string `json:"preview"`
	Name    string `json:"name"`
	Hd      string `json:"hd"`
}

type VostInfo struct {
	Data []VostInfoData `json:"data"`
}

type VostInfoData struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	Year  string `json:"year"`
	Genre string `json:"genre"`
}

func getStationsVost(body []byte) ([]VostPlaylist, error) {
	var s []VostPlaylist
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("Error parse json:", err)
	}
	return s, err
}

type nilbody struct{}

func (nilbody) Read([]byte) (int, error) {
	return 0, io.EOF
}

func testUrlVost(i VostPlaylist) (string, string) {
	body := nilbody{}
	req, err := http.NewRequest("GET", i.Hd, body)
	if err != nil {
		log.Fatal(err)
	}
	req.ContentLength = 0
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return i.Hd, "HD"
	} else {
		return i.Std, "SD"
	}

}

func VostPlay(s []VostPlaylist, num int) {
	for _, i := range s {
		data := strings.Split(i.Name, " ")
		num2, _ := strconv.Atoi(data[0])
		if num2 == num && data[1] == "серия" {
			u, q := testUrlVost(i)
			log.Println(i.Name+",", "Качество:", q)
			mpvCmd := exec.Command("mpv", "--fullscreen", "--speed=1.8", u)
			_, err := mpvCmd.Output()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func VostGetInfo(code string) VostInfoData {
	resp, err := http.PostForm("https://api.animevost.org/v1/info", url.Values{"id": {code}})
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//fmt.Println(string(body))
	var vostinfo VostInfo
	err = json.Unmarshal(body, &vostinfo)
	if err != nil {
		fmt.Println("Error parse json:", err)
	}
	//fmt.Println(vostinfo.Data[0])
	return vostinfo.Data[0]
}

func vost(code string, startPtr int, mongoClient *mongo.Client) {
	var inf = VostGetInfo(code)
	fmt.Println("Аниме:", inf.Title, "\nГод:", inf.Year, "\nТип:", inf.Type, "\nЖанр:", inf.Genre, "\nСайт:", "Animevost")
	filter := bson.M{"id": code}
	var result VostDB
	coll := mongoClient.Database("Anime").Collection("vost")
	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		_, err := coll.InsertOne(context.TODO(), bson.M{"id": code, "start": 0})
		if err != nil {
			log.Fatalln(err)
		}
	}
	resp, err := http.PostForm("https://api.animevost.org/v1/playlist", url.Values{"id": {code}})
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	s, err := getStationsVost(body)
	var max = len(s) + 1
	var startFrom int
	if startPtr == 0 {
		startFrom = result.Start
	} else {
		startFrom = startPtr
	}
	for i := startFrom; i <= max; i++ {
		//noinspection GoNilness
		VostPlay(s, i)
		update := bson.D{
			{"$set", bson.D{
				{"start", i},
			}},
		}
		_, _ = coll.UpdateOne(context.TODO(), filter, update)
	}

}
