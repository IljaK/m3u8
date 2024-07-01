package meta

import (
	"bufio"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"m3u8/cfg"
	"m3u8/util"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Record struct {
	GroupName string // #EXTGRP:HD
	NameData  string // #EXTINF:0,Россия HD / #EXTINF:10.000000,
	Url       string
}

func (r *Record) IsFilled() bool {
	return r.Url != "" && r.NameData != ""
}

type multiGroup struct {
	mux             sync.Mutex
	lowResChannels  []*Channel
	highResChannels []*Channel
}

func (m *multiGroup) AddChannel(chnl *Channel) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if chnl.Width >= 1920 || chnl.Height >= 1080 {
		m.highResChannels = append(m.highResChannels, chnl)
		return
	}
	m.lowResChannels = append(m.lowResChannels, chnl)
}

func (m *multiGroup) Contains(channelName string) bool {
	m.mux.Lock()
	defer m.mux.Unlock()

	for _, channel := range m.lowResChannels {
		if channel.Name == channelName {
			return true
		}
	}
	for _, channel := range m.highResChannels {
		if channel.Name == channelName {
			return true
		}
	}
	return false
}

type Media struct {
	forceReloadChannelData bool
	noSampleLoad           bool
	validFileType          bool

	Version        string // #EXT-X-VERSION:3
	MediaSequence  string // #EXT-X-MEDIA-SEQUENCE:20456
	TargetDuration string // #EXT-X-TARGETDURATION:11

	Records []*Record

	// Groups with channel names
	Groups []*Group
}

func (m *Media) lastRecord() *Record {
	if len(m.Records) == 0 {
		return nil
	}
	return m.Records[len(m.Records)-1]
}

func (m *Media) addGroup(record *Record) {
	if !record.IsFilled() {
		return
	}
	var group *Group
	for i := 0; i < len(m.Groups); i++ {
		if m.Groups[i].Name == record.GroupName {
			group = m.Groups[i]
		}
	}
	if group == nil {
		group = &Group{
			Name:     record.GroupName,
			Channels: nil,
		}
		m.Groups = append(m.Groups, group)
	}

	channel := Channel{
		Url:             record.Url,
		ForceReloadData: m.forceReloadChannelData,
		NoSampleLoad:    m.noSampleLoad,
	}
	channel.SetName(record.NameData, record.GroupName)

	group.Channels = append(group.Channels, &channel)
}
func (m *Media) CreateGroup(name string) *Group {
	group, _ := m.FindGroup(name)
	if group == nil {
		group = &Group{
			Name:     name,
			Channels: nil,
		}
		m.Groups = append(m.Groups, group)
	}
	return group
}

func (m *Media) FindGroup(name string) (*Group, int) {
	for i := 0; i < len(m.Groups); i++ {
		if m.Groups[i] != nil && m.Groups[i].Name == name {
			return m.Groups[i], i
		}
	}
	return nil, -1
}

func (m *Media) AddLine(line string) error {

	if strings.HasPrefix(line, "#EXTM3U") {
		m.validFileType = true
		// We are at begin of processing file
		return nil
	}
	if !m.validFileType {
		return errors.New("invalid file type with first line: " + line)
	}

	if strings.HasPrefix(line, "#EXT-X-VERSION:") {
		m.Version = line[len("#EXT-X-VERSION:"):]
		return nil
	}
	if strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:") {
		m.MediaSequence = line[len("#EXT-X-MEDIA-SEQUENCE:"):]
		return nil
	}
	if strings.HasPrefix(line, "#EXT-X-TARGETDURATION:") {
		m.TargetDuration = line[len("#EXT-X-TARGETDURATION:"):]
		return nil
	}

	var record = m.lastRecord()

	if record == nil || record.IsFilled() {
		record = &Record{}
		m.Records = append(m.Records, record)
	}

	// Record tags
	if strings.HasPrefix(line, "#EXTINF:") {
		// #EXTINF:0,Первый HD
		record.NameData = line[len("#EXTINF:"):]
	} else if strings.HasPrefix(line, "#EXTGRP:") {
		// #EXTGRP:HD
		record.GroupName = line[len("#EXTGRP:"):]
	} else if strings.HasPrefix(line, "#") {
		log.Println("Unknown tag: " + line)
	} else {
		record.Url = line
	}

	return nil
}

func (m *Media) SortGroup(groupName string) {
	group, _ := m.FindGroup(groupName)
	if group == nil {
		return
	}
	group.sortChannels()
}

func (m *Media) CheckHighRes(groupName string, fullSearch bool, threads int) {
	lowResGroup, _ := m.FindGroup(groupName)
	if lowResGroup == nil {
		return
	}

	highResGroupName := groupName + " HD"
	hiResGroup, _ := m.FindGroup(highResGroupName)
	if hiResGroup == nil {
		// Add group?
		hiResGroup = m.CreateGroup(highResGroupName)
	}
	var separated multiGroup
	if fullSearch {
		lowResGroup.mergeChannels(hiResGroup)
	} else {
		separated.highResChannels = append(separated.highResChannels, hiResGroup.Channels...)
	}

	for _, channel := range lowResGroup.Channels {
		separated.AddChannel(channel)
	}

	lowResGroup.Channels = separated.lowResChannels
	hiResGroup.Channels = separated.highResChannels
}

func ReadUrl(url string, forceReloadChannelData bool, noSampleLoad bool) *Media {

	http.DefaultClient.Timeout = 10 * time.Second
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to load playlist %s metadata: %v\n", url, err)
		return nil
	}
	if resp == nil || resp.Body == nil {
		log.Printf("Failed to load playlist %s metadata: zero response\n", url)
		return nil
	}

	var media *Media
	media, err = readRecords(resp.Body)
	media.forceReloadChannelData = forceReloadChannelData
	media.noSampleLoad = noSampleLoad
	_ = resp.Body.Close()

	if err != nil {
		log.Printf("Failed to read url %s data: %v\n", url, err)
		return nil
	}

	if media != nil {
		media.structRecords()
	}

	return media
}

func readRecords(r io.Reader) (*Media, error) {
	var media Media

	sc := bufio.NewScanner(r)
	var err error

	for sc.Scan() {
		err = media.AddLine(sc.Text())
		if err != nil {
			return &media, err
		}
	}

	return &media, sc.Err()
}

func (m *Media) structRecords() {
	for _, record := range m.Records {
		if record.IsFilled() {
			if record.GroupName != "" {
				// We found channel group
				m.addGroup(record)
			}
		}
	}
}

func (m *Media) WriteFiles(filePaths []string, epgUrl string, skipGroups []string) {
	for _, path := range filePaths {
		if len(path) == 0 {
			continue
		}
		m.WriteFile(path, epgUrl, skipGroups)
	}
}

func (m *Media) WriteFile(filePath string, epgUrl string, skipGroups []string) {
	if filePath == "" {
		log.Errorf("empty file path")
		return
	}

	f, err := os.Create(filePath)

	if f != nil {
		defer f.Close()
	}

	if err != nil {
		log.Errorf("failed to open file: %s with error: %+v", filePath, err)
		return
	}

	if epgUrl == "" {
		_, err = f.WriteString("#EXTM3U\n")
	} else {
		_, err = f.WriteString("#EXTM3U x-tvg-url=\"" + epgUrl + "\"\n")
	}

	if err != nil {
		log.Errorf("failed to write to file: %s with error: %+v", filePath, err)
		return
	}

	for _, group := range m.Groups {
		if util.Contains(skipGroups, group.Name) {
			continue
		}

		censored := strings.Contains(group.Name, "взрослые")

		for _, channel := range group.Channels {

			//  #EXTINF:0,Первый HD
			_, err = f.WriteString(channel.GetInfoData(censored) + "\n")
			if err != nil {
				log.Errorf("failed to write to file: %s with error: %+v", filePath, err)
				return
			}
			// #EXTGRP:HD
			_, err = f.WriteString("#EXTGRP:" + group.Name + "\n")
			if err != nil {
				log.Errorf("failed to write to file: %s with error: %+v", filePath, err)
				return
			}
			// URL
			_, err = f.WriteString(channel.Url + "\n")

			if err != nil {
				log.Errorf("failed to write to file: %s with error: %+v", filePath, err)
				return
			}
		}
	}
	log.Infof("Wrote %s", filePath)
}
func (m *Media) forceChannels(groupName string, channelNames []string) {
	if groupName == "" {
		return
	}
	if channelNames == nil {
		return
	}
	if len(channelNames) == 0 {
		return
	}
	group := m.CreateGroup(groupName)
	if group == nil {
		return
	}

	channels := make([]*Channel, 0, len(channelNames))

	for _, chnl := range channelNames {
		for _, g := range m.Groups {
			if g.Name != group.Name {
				for i := len(g.Channels) - 1; i >= 0; i-- {
					channel := g.Channels[i]
					if strings.ToLower(channel.Name) == strings.ToLower(chnl) {
						// No breaking here, there can be multiple channels with same name!
						g.extractChannel(i)
						channels = append(channels, channel)
					}
				}
			}
		}
	}
	group.Channels = append(group.Channels, channels...)

}

func (m *Media) ApplyGroupsForcing() {

	groupsConf := cfg.GetGroups()

	for _, item := range groupsConf {

		switch item.(type) {
		case map[string]interface{}:
			name := util.GetValue("name", item.(map[string]interface{}), "")

			force := util.GetValueArray("force", item.(map[string]interface{}), []string{})
			begin := util.GetValueArray("begin", item.(map[string]interface{}), []string{})
			end := util.GetValueArray("end", item.(map[string]interface{}), []string{})

			force = append(force, begin...)
			force = append(force, end...)

			m.forceChannels(name, force)

			break
		}
	}
}

func (m *Media) FindChannel(channelName string) (*Group, *Channel, int) {
	for _, group := range m.Groups {

		for i2, channel := range group.Channels {
			if channel.Name == channelName {
				return group, channel, i2
			}
		}
	}
	return nil, nil, 0
}

func (m *Media) FilterForeign(groupName string) {

	group, _ := m.FindGroup(groupName)
	if group == nil {
		return
	}

	reg, err := regexp.Compile("(?i)( (AU|BR|DE|PL|QA|PT|SE|FR|UA|LV|SK|IN|AE|ES|LT|AZ|KG|RO|DK|NO|IT|BE|HR|FI|NL|BG|CZ|MD|LA|EG|IE|ER|DC))$")
	if err != nil {
		log.Fatal(err)
		return
	}

	foreignGroup := m.CreateGroup("иностранные")

	for n := len(group.Channels) - 1; n >= 0; n-- {
		last := len(group.Channels) - 1
		channel := group.Channels[n]
		if reg.Match([]byte(channel.Name)) {
			foreignGroup.Channels = append(foreignGroup.Channels, channel)
			group.Channels[n] = group.Channels[last]
			group.Channels = group.Channels[:last]
		}
	}
	/*

		// To detect all
		result := make([]string, 0, 10)

		for _, group := range m.Groups {
			for _, channel := range group.Channels {
				words := strings.Split(channel.Name, " ")
				if len(words) > 1 {
					last := words[len(words)-1]
					if len(last) == 2 && !util.ContainsString(result, last) {
						result = append(result, last)
					}
				}
			}
		}
		log.Printf("result: %v", result)
	*/
}

func (m *Media) OrderGroups() {
	order := cfg.GetGroupOrder()
	if order == nil || len(order) == 0 {
		return
	}

	ordered := make([]*Group, 0, len(m.Groups))

	for _, s := range order {
		group, index := m.FindGroup(s)
		if index >= 0 {
			ordered = append(ordered, group)
			m.Groups[index] = nil
		}
	}

	for _, group := range m.Groups {
		if group != nil {
			ordered = append(ordered, group)
		}
	}
	m.Groups = ordered
}

func (m *Media) PrintGroups() {
	log.Print("Groups: [")
	for i, group := range m.Groups {
		if i > 0 {
			log.Print(" ")
		}
		log.Printf("'%s'", group.Name)
	}
	log.Println("]")
}

func (m *Media) ValidateHighRes() {
	validationList := make([]string, 0, 10)

	groupsConf := cfg.GetHDSplit()

	for _, g := range groupsConf {
		group, _ := m.FindGroup(g)
		if group == nil {
			continue
		}

		if !strings.Contains(group.Name, "HD") && !strings.Contains(group.Name, "4K") {
			validationList = append(validationList, group.Name)
		}
	}
	for _, groupName := range validationList {
		m.CheckHighRes(groupName, true, 2)
	}
}

func (m *Media) SortGroups() {
	for _, group := range m.Groups {
		group.sortChannels()
	}
}
func (m *Media) FindRecord(url string) *Record {
	for _, record := range m.Records {
		if record.Url == url {
			return record
		}
	}
	return nil
}
