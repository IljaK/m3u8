package main

import (
	log "github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"m3u8/cfg"
	"m3u8/db"
	"m3u8/meta"
	"m3u8/util"
	"sync"
)

func processChannels(media *meta.Media) {

	//media.PrintGroups()

	media.ApplyGroupsForcing()
	media.ValidateHighRes()
	media.SortGroups()
	media.OrderGroups()
}

func processFile(wg *sync.WaitGroup, file fs.FileInfo) {
	defer wg.Done()
	media := meta.ReadFile("input/" + file.Name())

	if media == nil {
		return
	}

	processChannels(media)
	media.WriteFile("output/" + file.Name())
}

func readInputFiles() {
	files, err := ioutil.ReadDir("input/")
	if err != nil {
		log.Fatal(err)
		return
	}

	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go processFile(&wg, file)
	}
	wg.Wait()
}

func compareResult(fileName string) {
	mediaIn := meta.ReadFile("input/" + fileName)
	if mediaIn == nil {
		return
	}

	mediaOut := meta.ReadFile("output/" + fileName)
	if mediaOut == nil {
		return
	}

	for _, record := range mediaIn.Records {
		if mediaOut.FindRecord(record.Url) == nil {
			log.Printf("Missing:\n%s\n%s\n%s\n\n", record.NameData, record.GroupName, record.Url)
		}
	}
}

func loadPlayList(url string, providerType string) {
	if url == "" {
		log.Errorf("invalid url in list")
		return
	}
	if providerType == "" {
		log.Errorf("invalid providerType in list")
		return
	}

	media := meta.ReadUrl(url, providerType)

	if media == nil {
		return
	}
	processChannels(media)
	media.WriteFile("output/" + providerType + ".m3u8")
}

func processListConfig() {
	wg := sync.WaitGroup{}

	lists := cfg.GetLists()
	if lists == nil {
		log.Println("Empty lists config")
		return
	}

	for _, item := range lists {
		switch item.(type) {
		case map[string]interface{}:
			wg.Add(1)
			processList(&wg, item.(map[string]interface{}))
		default:
			break
		}
	}
	wg.Wait()
}

func processList(wg *sync.WaitGroup, cfg map[string]interface{}) {
	defer wg.Done()
	loadPlayList(util.GetValue("url", cfg, ""), util.GetValue("type", cfg, ""))
}

func main() {

	cfg.LoadConfig()
	err := db.Init("postgres://iljakrusman:iljakrusman@127.0.0.1:5432/m3u8?sslmode=disable")
	if err != nil {
		panic(err)
	}
	log.Println("Processing data...")
	processListConfig()
	//readInputFiles()
	// compareResult("edem_playlist.m3u8");
	log.Println("Completed!")
}
