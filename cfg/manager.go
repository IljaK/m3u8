package cfg

import (
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"m3u8/util"
	"os"
	"strings"
)

var conf map[string]interface{}
var viperEnv *viper.Viper

//var Conf *viper.Viper

func LoadConfig(configFile string, envFile string) error {

	viperEnv = viper.New()
	viperEnv.SetConfigType("env")
	viperEnv.SetConfigFile(envFile)

	viperEnv.AutomaticEnv()

	err := viperEnv.ReadInConfig()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configFile)

	if err != nil {
		return fmt.Errorf("failed to open file %s with error: %+v", configFile, err)
	}

	err = yaml.Unmarshal(data, &conf)

	if err != nil {
		return fmt.Errorf("failed to unmarshall yaml file %s with error: %+v", configFile, err)
	}

	return nil
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
