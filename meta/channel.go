package meta

import (
	"log"
	"m3u8/db"
	"m3u8/ffprobe"
	"m3u8/util"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Channel struct {
	Name        string
	SortingName string
	InfoData    string
	Url         string

	HistoryDays int
	Width       int
	Height      int

	ProviderType string

	meta *Media
}

func (c *Channel) setData(data string) {
	c.InfoData = data

	// #EXTINF:0 tvg-rec="0",минимакс-воронины HD
	variables := strings.Split(data, " ")
	for _, variableRaw := range variables {
		key, val := util.ParseVariable(variableRaw)
		if key == "tvg-rec" {
			days, _ := strconv.ParseInt(val, 10, 32)
			c.HistoryDays = int(days)
		}
	}
}

func (c *Channel) SetName(nameData string, groupName string) {

	content := strings.SplitN(nameData, ",", 2)

	if content != nil && len(content) > 0 {
		c.setData(content[0])
		if len(content) > 1 {
			c.Name = content[1]
		}
	}

	reg, err := regexp.Compile("^([0-9]|\\t| |\\.)+")
	if err != nil {
		log.Fatal(err)
	}
	c.SortingName = strings.ToLower(reg.ReplaceAllString(c.Name, ""))

	u, err := url.Parse(c.Url)
	splittedPath := strings.Split(u.Path, "/")
	if len(splittedPath) < 3 {
		log.Println("Error in url path:", splittedPath)
		return
	}
	remoteId := splittedPath[3]

	//if remoteId == "16140" {
	//	log.Println("Test!")
	//}

	channelData, err := db.QueryGetChannelInfo(remoteId)

	if channelData == nil {
		c.loadMeta()

		err = db.QueryAddChannelInfo(&db.Channel{
			Id:          0,
			Name:        c.Name,
			RemoteId:    remoteId,
			Width:       c.Width,
			Height:      c.Height,
			HistoryDays: c.HistoryDays,
			Group:       groupName,
		})
		if err != nil {
			log.Println(err)
		}
	} else {
		if channelData.Width == 0 || channelData.Height == 0 {
			c.loadMeta()
			channelData.Width = c.Width
			channelData.Height = c.Height
		} else {
			c.Width = channelData.Width
			c.Height = channelData.Height
		}

		err = db.QueryUpdateChannel(channelData)
		if err != nil {
			log.Println(err)
		}
		err = db.QueryUpdateProvider(channelData, c.ProviderType)
		if err != nil {
			log.Println(err)
		}
	}

}

func (c *Channel) loadMeta() *ffprobe.MetaData {
	if c.Url == "" {
		return nil
	}

	media := ReadUrl(c.Url, c.ProviderType)

	if media != nil && len(media.Records) > 0 {

		var metaData *ffprobe.MetaData
		for i := len(media.Records) - 1; i >= 0; i-- {
			metaData = ffprobe.LoadMetaData(media.Records[i].Url)
			if metaData != nil {
				vidStream := metaData.GetVideoStream()
				if vidStream != nil && vidStream.Width != 0 && vidStream.Height != 0 {
					c.Width = vidStream.Width
					c.Height = vidStream.Height
					return metaData
				}
			}
		}
		return metaData
	}
	return nil
}
