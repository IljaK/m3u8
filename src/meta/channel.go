package meta

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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
	TvgName     string
	infoData    string
	Url         string

	HistoryDays int
	Width       int
	Height      int
	FrameRate   int

	//providerHost string
	//providerName string

	ForceReloadData bool
	NoSampleLoad    bool

	meta *Media
}

/*
func (c *Channel) GetProviderHost() string {
	if c.providerHost != "" {
		return c.providerHost
	}
	if c.Url == "" {
		return ""
	}
	values, err := url.Parse(c.Url)

	if err != nil {
		return ""
	}

	args := strings.Split(values.Host, ".")
	if len(args) > 1 {
		return strings.Join(args[1:], ".")
	}
	return strings.Join(args, ".")
}
*/

func (c *Channel) setData(data string) {
	c.infoData = data

	// #EXTINF: 0 catchup="default" catchup-days="5", Disney Channel
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

func (c *Channel) GetInfoData(censored bool) string {
	// #EXTINF: 0 catchup="default" catchup-days="5",Disney Channel
	// #EXTINF:0 tvg-rec="0",минимакс-воронины HD
	// censored=1

	result := fmt.Sprintf("#EXTINF:0 tvg-rec=\"%d\" catchup=\"shift\" catchup-days=\"%d\"", c.HistoryDays, c.HistoryDays)
	if c.TvgName != "" {
		result += fmt.Sprintf(" tvg-name=\"%s\"", c.TvgName)
	}
	if censored {
		result += " censored=\"1\""
	}
	return result + "," + c.Name
}

func (c *Channel) SetName(nameData string, groupName string) {

	content := strings.SplitN(nameData, ",", 2)

	if content != nil && len(content) > 0 {
		c.setData(content[0])
		if len(content) > 1 {
			c.Name = strings.TrimSpace(content[1])
		}
	}

	reg, err := regexp.Compile(`(^([0-9]+))|(\.|\+|-|\s|,|_)`)
	if err != nil {
		log.Fatal(err)
	}
	c.SortingName = strings.ToLower(reg.ReplaceAllString(c.Name, ""))

	// http://wkejhfk.rossteleccom.net/iptv/ABCD3HG7DW38ZD/205/index.m3u8
	// host + / + "iptv" + / + key + / + channel_id + / + file

	u, err := url.Parse(c.Url)
	splittedPath := strings.Split(u.Path, "/")
	if len(splittedPath) < 3 {
		log.Println("Error in url path:", splittedPath)
		return
	}
	remoteId := splittedPath[3]
	provider := db.Provider{}
	provider.FromUri(u.Host, splittedPath)

	channelData, err := db.QueryGetChannelInfo(remoteId, &provider)

	if channelData == nil || ((!c.NoSampleLoad && !channelData.HasAllMeta()) || c.ForceReloadData) {
		if c.loadMeta(remoteId) == nil {
			log.Printf("Failed to load channel meta for remoteId: %s", remoteId)
		}
	} else {
		c.Width = channelData.Width
		c.Height = channelData.Height
		c.FrameRate = channelData.FrameRate
		c.TvgName = channelData.TvgName
	}

	if c.isNeedDBUpdate(channelData) || (channelData != nil && channelData.ChannelName.Group != groupName) {
		dbChannel := &db.Channel{
			Id:        0,
			RemoteId:  remoteId,
			Width:     c.Width,
			Height:    c.Height,
			FrameRate: c.FrameRate,
			ChannelName: db.ChannelName{
				Id:          0,
				Name:        c.Name,
				HistoryDays: c.HistoryDays,
				Group:       groupName,
				Provider:    provider,
			},
		}
		err = db.QueryInsertOrUpdateChannel(dbChannel)
		if err != nil {
			log.Println(err)
		}
	}
}

func (c *Channel) isNeedDBUpdate(dbChannel *db.Channel) bool {
	if dbChannel == nil {
		return true
	}
	if dbChannel.Id == 0 {
		return true
	}
	if dbChannel.Width != c.Width && c.Width != 0 {
		return true
	}
	if dbChannel.Height != c.Height && c.Height != 0 {
		return true
	}
	if dbChannel.FrameRate != c.FrameRate && c.FrameRate != 0 {
		return true
	}
	if dbChannel.ChannelName.Id == 0 {
		return true
	}
	if dbChannel.ChannelName.Name != c.Name {
		return true
	}
	if dbChannel.ChannelName.HistoryDays != c.HistoryDays {
		return true
	}
	if dbChannel.ChannelName.Provider.Id == 0 {
		return true
	}
	return false
}

func (c *Channel) loadMeta(remoteId string) *ffprobe.MetaData {
	if c.Url == "" {
		return nil
	}

	media := ReadUrl(c.Url, c.ForceReloadData, c.NoSampleLoad)

	if media != nil && len(media.Records) > 0 {

		var metaData *ffprobe.MetaData
		for i := len(media.Records) - 1; i >= 0; i-- {
			metaData = ffprobe.LoadMetaData(remoteId, media.Records[i].Url)
			if metaData != nil {
				vidStream := metaData.GetVideoStream()
				if vidStream != nil && vidStream.Width != 0 && vidStream.Height != 0 {
					c.Width = vidStream.Width
					c.Height = vidStream.Height
					c.FrameRate = vidStream.RFrameRate.RoundedQuotient()
					return metaData
				}
			}
		}
		return metaData
	}
	return nil
}
