package cmd

import (
	"github.com/spf13/cobra"
)

var ConfFile string
var EnvFile string
var LogFile string
var ForceReDownload bool
var NoSampleLoad bool
var NoTvGuide bool

var confCmd = &cobra.Command{
	Use:   "--conf=filepath -force",
	Short: "m3u8 is program for formatting huge channel list",
	Run:   nil,
}

func Init() error {
	confCmd.Flags().StringVarP(&ConfFile, "conf", "c", "./order.yaml", "order config file path")
	confCmd.Flags().StringVarP(&EnvFile, "env", "e", "./m3u8.env", "env file path")
	confCmd.Flags().StringVarP(&LogFile, "log", "l", "", "log file path")
	confCmd.Flags().BoolVarP(&ForceReDownload, "force", "f", false, "force reload channels dimensions")
	confCmd.Flags().BoolVarP(&NoSampleLoad, "no-sample", "s", false, "skip loading sample to update 0 size")
	confCmd.Flags().BoolVarP(&NoTvGuide, "no-tvg", "t", false, "skip including tv guide")
	return confCmd.Execute()
}
