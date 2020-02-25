package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AniPlaylist struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Sd    string `json:"sd"`
	Hd    string `json:"hd"`
	SrcHd string `json:"srcHd"`
	SrcSd string `json:"srcSd"`
}

type BlockedInfo struct {
	Blocked bool `json:"blocked"`
	Reason  bool `json:"reason"`
}

type Data struct {
	Id          int           `json:":id"`
	Series      string        `json:"series"`
	Names       []string      `json:"names"`
	Genres      []string      `json:"genres"`
	Poster      string        `json:"poster"`
	Status      string        `json:"status"`
	Type        string        `json:"type"`
	Year        string        `json:"year"`
	BlockedInfo BlockedInfo   `json:"blockedInfo"`
	Playlist    []AniPlaylist `json:"playlist"`
}

type Anilibria struct {
	Status bool `json:"status"`
	Data   Data `json:"data"`
}

func (s *Anilibria) PlayAnilibria(num int) {
	for _, i := range s.Data.Playlist {
		if i.Id == num {
			log.Println("Part:", i.Title, "APIPart", i.Id)
			mpvCmd := exec.Command("mpv", "--fullscreen", "--speed=1.8", i.Hd)
			_, err := mpvCmd.Output()
			if err != nil {
				fmt.Println(err)
			}
			//log.Println(string(mpvOut))
		}
	}
}
func getStations(body []byte) (*Anilibria, error) {
	var s = new(Anilibria)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("Error parse json:", err)
	}
	return s, err
}

func anilibria(code string, startPtr int, mongoClient *mongo.Client) {
	filter := bson.M{"code": code}
	var result AnilibriaDb
	coll := mongoClient.Database("Anime").Collection("anilibria")
	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		_, err := coll.InsertOne(context.TODO(), bson.M{"code": code, "start": 0})
		if err != nil {
			log.Fatalln(err)
		}
	}

	resp, err := http.PostForm("https://www.anilibria.tv/public/api/index.php", url.Values{"query": {"release"}, "code": {code}})
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	s, err := getStations(body)
	if err != nil {
		log.Fatalln(err)
	}
	//noinspection GoNilness

	fmt.Println("Аниме: "+s.Data.Names[0]+"\nГод: "+s.Data.Year+"\nТип: "+s.Data.Type+"\nЖанр: "+strings.Join(s.Data.Genres, ", "), "\nСайт: Anilibria.tv")
	var max = len(s.Data.Playlist) + 1
	var startFrom int
	if startPtr == 0 {
		startFrom = result.Start
	} else {
		startFrom = startPtr
	}
	for i := startFrom; i <= max; i++ {
		//noinspection GoNilness
		s.PlayAnilibria(i)
		update := bson.D{
			{"$set", bson.D{
				{"start", i},
			}},
		}
		_, _ = coll.UpdateOne(context.TODO(), filter, update)
	}
}
