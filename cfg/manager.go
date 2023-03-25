package cfg

import (
	"fmt"
	"github.com/spf13/viper"
	"m3u8/util"
	"strings"
)

var conf map[string]interface{}
var viperEnv *viper.Viper

//var Conf *viper.Viper

func LoadConfig(configFile string, envFile string) {
	vp := viper.New()
	vp.SetConfigFile(configFile)
	vp.SetConfigType("yaml")
	err := vp.ReadInConfig() // Find and read the config file
	if err != nil {          // Handle errors reading the config file
		panic(err)
	}
	err = vp.Unmarshal(&conf)
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("config read error: %v", err))
	}

	viperEnv = viper.New()
	viperEnv.SetConfigType("env")
	viperEnv.SetConfigFile(envFile)

	viperEnv.AutomaticEnv()

	err = viperEnv.ReadInConfig()
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
func GetTvGuide() map[string]string {
	return util.GetValueMap("tvguide", conf, map[string]string{})
}
func GetEnvString(key string, defVal string) string {
	key = strings.ToLower(key)
	viperEnv.GetString(key)
	if !viperEnv.IsSet(key) {
		return defVal
	}

	return viperEnv.GetString(key)
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
