package cfg

import "m3u8/util"

type Output struct {
	FileName   string
	SkipGroups []string
}

func (l *Output) Load(cfg map[string]interface{}) {
	l.FileName = util.GetValue("file_name", cfg, "")
	l.SkipGroups = util.GetValueArray("skip_groups", cfg, []string{})
}

type List struct {
	Url     string
	EpgUrl  string
	Outputs []Output
}

func (l *List) Load(cfg map[string]interface{}) {
	l.Url = util.GetValue("url", cfg, "")
	l.EpgUrl = util.GetValue("epg_url", cfg, "")

	outputs := util.GetValueArray("output", cfg, []map[string]interface{}{})
	l.Outputs = make([]Output, len(outputs), len(outputs))

	for i := 0; i < len(outputs); i++ {
		l.Outputs[i].Load(outputs[i])
	}

	//l.Output = util.GetValueArray("output", cfg, []string{})
	//l.SkipGroups = util.GetValueArray("skip_groups", cfg, []string{})
}

func Load(cfg map[string]interface{}) *List {
	l := List{}
	l.Load(cfg)
	return &l
}
