package cfg

import (
	"fmt"
	"github.com/spf13/viper"
	"m3u8/util"
)

var conf map[string]interface{}

//var Conf *viper.Viper

func LoadConfig() {
	vp := viper.New()
	vp.SetConfigName("order")
	vp.SetConfigType("yaml")
	vp.AddConfigPath(".")
	err := vp.ReadInConfig() // Find and read the config file
	if err != nil {          // Handle errors reading the config file
		panic(err)
	}
	err = vp.Unmarshal(&conf)
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("config read error: %v", err))
	}
}

func GetGroups() []interface{} {
	return util.GetValueArray("groups", conf, []interface{}{})
}

func GetLists() []interface{} {
	return util.GetValueArray("lists", conf, []interface{}{})
}

func GetHDSplit() []string {
	return util.GetValueArray("group_hd_split", conf, []string{})
}
func GetGroupOrder() []string {
	return util.GetValueArray("group_order", conf, []string{})
}

func GetGroupConfig(groupName string) map[string]interface{} {
	groupsConf := GetGroups()

	for _, item := range groupsConf {

		switch item.(type) {
		case map[string]interface{}:
			name := util.GetValue("name", item.(map[string]interface{}), "")
			if name == groupName {
				return item.(map[string]interface{})
			}
			break
		}
	}
	return map[string]interface{}{}
}
