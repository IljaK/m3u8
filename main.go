package main

import (
	log "github.com/sirupsen/logrus"
	"m3u8/cfg"
	"m3u8/cmd"
	"m3u8/db"
	"m3u8/meta"
	"m3u8/util"
	"sync"
)

func processChannels(media *meta.Media) {

	//media.PrintGroups()

	// Old way:
	//media.ApplyGroupsForcing()
	//media.ValidateHighRes()
	//media.SortGroups()
	//media.OrderGroups()

	media.ApplyGroupsForcing()
	media.SortGroups()
	media.ValidateHighRes()
	media.OrderGroups()
}

func loadPlayList(url string, output string, forceReloadChannelData bool) {
	if url == "" {
		log.Errorf("invalid url in list")
		return
	}
	if output == "" {
		log.Errorf("invalid output list")
		return
	}

	media := meta.ReadUrl(url, forceReloadChannelData)

	if media == nil {
		return
	}
	processChannels(media)
	media.WriteFile(output)
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
	loadPlayList(util.GetValue("url", cfg, ""),
		util.GetValue("output", cfg, ""),
		cmd.ForceReDownload)
}

func main() {

	err := cmd.Init()
	if err != nil {
		panic(err)
	}

	cfg.LoadConfig(cmd.ConfPath, cmd.EnvPath)

	err = db.Init(cfg.GetEnvString("DB_URI", ""))
	if err != nil {
		panic(err)
	}
	log.Println("Processing data...")
	processListConfig()
	//readInputFiles()
	// compareResult("edem_playlist.m3u8");
	log.Println("Completed!")
}
