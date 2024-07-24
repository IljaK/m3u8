package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"m3u8/cfg"
	"m3u8/cmd"
	"m3u8/db"
	"m3u8/meta"
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

func loadPlayList(data *cfg.List, forceReloadChannelData bool, noSampleLoad bool) {
	if data.Url == "" {
		log.Errorf("invalid url in list")
		return
	}

	media := meta.ReadUrl(data.Url, forceReloadChannelData, noSampleLoad)

	if media == nil {
		return
	}
	processChannels(media)

	for _, output := range data.Outputs {
		media.WriteFile(output.FileName, data.EpgUrl, output.SkipGroups)
	}
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
			go processList(&wg, item.(map[string]interface{}))
		default:
			break
		}
	}
	wg.Wait()
}

func processList(wg *sync.WaitGroup, cfgMap map[string]interface{}) {
	defer wg.Done()

	loadPlayList(cfg.Load(cfgMap), cmd.ForceReDownload,
		cmd.NoSampleLoad)
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

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	err := cmd.Init()
	if err != nil {
		panic(err)
	}

	setupLog(cmd.LogFile)

	must(cfg.LoadConfig(cmd.ConfFile, cmd.EnvFile))

	must(db.Init(cfg.GetEnvString("DB_URI", "")))

	if !cmd.NoTvGuide {
		err = xmltv.GenerateTvGuideFromUrl(cfg.GetTvGuide())
		if err != nil {
			log.Errorf("Failed to generate Tv Guide")
		}
	}

	log.Println("Processing play lists...")
	processListConfig()
	log.Println("Completed!")

	// Wait till all async DB Queries complete
	// TODO: timeout
	db.WaitAllComplete()
}
