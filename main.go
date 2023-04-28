package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"m3u8/cfg"
	"m3u8/cmd"
	"m3u8/db"
	"m3u8/meta"
	"m3u8/util"
	"m3u8/xmltv"
	"os"
	"sync"
)

func processChannels(media *meta.Media) {
	media.ApplyGroupsForcing()
	media.SortGroups()
	media.ValidateHighRes()
	media.OrderGroups()
}

func loadPlayList(url string, output string, epgUrl string, forceReloadChannelData bool, noSampleLoad bool) {
	if url == "" {
		log.Errorf("invalid url in list")
		return
	}
	if output == "" {
		log.Errorf("invalid output list")
		return
	}

	media := meta.ReadUrl(url, forceReloadChannelData, noSampleLoad)

	if media == nil {
		return
	}
	processChannels(media)
	media.WriteFile(output, epgUrl)
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
	loadPlayList(util.GetValue("url", cfg, ""), util.GetValue("output", cfg, ""),
		util.GetValue("epg_url", cfg, ""), cmd.ForceReDownload, cmd.NoSampleLoad)
}

func setupLog(filePath string) {

	if filePath != "" {
		log.SetFormatter(&log.JSONFormatter{})
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			return
		}
		log.SetOutput(io.Writer(f))
		//mw := io.MultiWriter(os.Stdout, f)
		//log.SetOutput(mw)
	}
}

func main() {

	err := cmd.Init()
	if err != nil {
		panic(err)
	}

	setupLog(cmd.LogFile)

	cfg.LoadConfig(cmd.ConfFile, cmd.EnvFile)

	err = db.Init(cfg.GetEnvString("DB_URI", ""))
	if err != nil {
		panic(err)
	}
	if !cmd.NoTvGuide {
		err = xmltv.GenerateTvGuideFromUrl(cfg.GetTvGuide())
		if err != nil {
			log.Errorf("Failed to generate Tv Guide")
		}
	}

	log.Println("Processing play lists...")
	processListConfig()
	log.Println("Completed!")
}
