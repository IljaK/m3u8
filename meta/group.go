package meta

import (
	"log"
	"m3u8/util"
	"sort"
	"strings"
)

type Group struct {
	Name     string
	Channels []*Channel
}

func (g *Group) FindChannel(channelName string) (*Channel, int) {
	for i2, channel := range g.Channels {
		if channel != nil && strings.ToLower(channel.Name) == strings.ToLower(channelName) {
			return channel, i2
		}
	}
	return nil, -1
}

func (g *Group) mergeChannels(group *Group) {
	g.Channels = append(g.Channels, group.Channels...)
	group.Channels = make([]*Channel, 0, len(group.Channels))
}

func (g *Group) sortChannels() {

	groupConf := g.getConfig()

	begin := util.GetStringArrayKey("begin", groupConf)
	end := util.GetStringArrayKey("end", groupConf)

	beginChannels := make([]*Channel, 0, len(begin))
	endChannels := make([]*Channel, 0, len(end))

	if begin != nil {
		for i := 0; i < len(begin); i++ {
			_, targetIndex := g.FindChannel(begin[i])
			if targetIndex >= 0 {
				beginChannels = append(beginChannels, g.extractChannel(targetIndex))
			}
		}
	}

	if end != nil {
		for i := 0; i < len(end); i++ {
			_, targetIndex := g.FindChannel(end[i])
			if targetIndex >= 0 {
				endChannels = append(endChannels, g.extractChannel(targetIndex))
			}
		}
	}

	sort.Slice(g.Channels, func(i int, j int) bool {
		return g.Channels[i].SortingName < g.Channels[j].SortingName
	})

	g.Channels = append(beginChannels, g.Channels...)
	g.Channels = append(g.Channels, endChannels...)
}

func (g *Group) PrintChannels() {

	log.Println(g.Name)

	log.Print("[")
	for i, channel := range g.Channels {
		if i != 0 {
			log.Printf(", ")
		}
		log.Printf("'%s'", channel.Name)
	}
	log.Print("]\n")
}

func (g *Group) extractChannel(index int) *Channel {
	var chnl *Channel = nil

	if index < 0 {
		return chnl
	}
	if index >= len(g.Channels) {
		return chnl
	}

	last := len(g.Channels) - 1
	chnl = g.Channels[index]
	if index < last {
		g.Channels[index] = g.Channels[last]
	}
	g.Channels = g.Channels[:last]

	return chnl
}

func (g *Group) getConfig() map[string]interface{} {
	groupsConf := Conf.Get("groups")
	if groupsConf == nil {
		return nil
	}

	for _, item := range groupsConf.([]interface{}) {
		switch item.(type) {
		case map[string]interface{}:
			vals := item.(map[string]interface{})
			grpName := vals["name"]
			if grpName == g.Name {
				return vals
			}
			break
		default:
			break
		}
	}
	return nil
}
