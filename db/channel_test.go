package db

import (
	"m3u8/cfg"
	"m3u8/cmd"
	"os"
	"path"
	"runtime"
	"testing"
)

func initDB(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		t.Fatalf("initDB err: %v", err)
	}

	err = cmd.Init()
	if err != nil {
		t.Fatalf("cmd.Init err: %v", err)
	}

	_ = cfg.LoadConfig(cmd.ConfFile, cmd.EnvFile)

	//if err != nil {
	//	t.Fatalf("cfg.LoadConfig err: %v", err)
	//}

	err = Init(cfg.GetEnvString("DB_URI", ""))

	if err != nil {
		t.Fatalf("Init err: %v", err)
	}
}

func TestQueryAddOrUpdateChannelName(t *testing.T) {

	initDB(t)

	channel := Channel{
		RemoteId:    "1234",
		Width:       1280,
		Height:      720,
		FrameRate:   60,
		ChannelName: ChannelName{},
	}

	err := QueryInsertOrUpdateChannel(&channel)

	if err != nil {
		t.Fatalf("Error inserting or updating channel: %v", err)
	}
	WaitAllComplete()

}

func TestQueryGetChannelInfo(t *testing.T) {

	initDB(t)

	provider := Provider{
		Id:        0,
		Name:      "",
		Host:      "",
		SubDomain: "",
		AccessKey: "",
	}

	remoteId := "407"

	c, err := QueryGetChannelInfo(remoteId, &provider)

	if err != nil {
		t.Fatalf("Failed to QueryGetChannelInfo DB: %v", err)
	}

	if c == nil {
		t.Fatalf("Failed to get channel info")
	}

	WaitAllComplete()
}
