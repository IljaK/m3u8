package ffprobe

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	utils "m3u8/util"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
	{
	    "streams": [
	        {
	            "index": 0,
	            "codec_name": "h264",
	            "codec_long_name": "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10",
	            "profile": "High",
	            "codec_type": "video",
	            "codec_tag_string": "[27][0][0][0]",
	            "codec_tag": "0x001b",
	            "width": 720,
	            "height": 576,
	            "coded_width": 720,
	            "coded_height": 576,
	            "closed_captions": 0,
	            "film_grain": 0,
	            "has_b_frames": 0,
	            "sample_aspect_ratio": "64:45",
	            "display_aspect_ratio": "16:9",
	            "pix_fmt": "yuv420p",
	            "level": 30,
	            "chroma_location": "left",
	            "field_order": "progressive",
	            "refs": 1,
	            "is_avc": "false",
	            "nal_length_size": "0",
	            "ts_id": "1",
	            "ts_packetsize": "188",
	            "id": "0xd3",
	            "r_frame_rate": "25/1",
	            "avg_frame_rate": "25/1",
	            "time_base": "1/90000",
	            "start_pts": 1004577190,
	            "start_time": "11161.968778",
	            "duration_ts": 900000,
	            "duration": "10.000000",
	            "bits_per_raw_sample": "8",
	            "extradata_size": 46,
	            "disposition": {
	                "default": 0,
	                "dub": 0,
	                "original": 0,
	                "comment": 0,
	                "lyrics": 0,
	                "karaoke": 0,
	                "forced": 0,
	                "hearing_impaired": 0,
	                "visual_impaired": 0,
	                "clean_effects": 0,
	                "attached_pic": 0,
	                "timed_thumbnails": 0,
	                "non_diegetic": 0,
	                "captions": 0,
	                "descriptions": 0,
	                "metadata": 0,
	                "dependent": 0,
	                "still_image": 0
	            }
	        },
	        {
	            "index": 1,
	            "codec_name": "aac",
	            "codec_long_name": "AAC (Advanced Audio Coding)",
	            "profile": "LC",
	            "codec_type": "audio",
	            "codec_tag_string": "[15][0][0][0]",
	            "codec_tag": "0x000f",
	            "sample_fmt": "fltp",
	            "sample_rate": "44100",
	            "channels": 2,
	            "channel_layout": "stereo",
	            "bits_per_sample": 0,
	            "initial_padding": 0,
	            "ts_id": "1",
	            "ts_packetsize": "188",
	            "id": "0xdd",
	            "r_frame_rate": "0/0",
	            "avg_frame_rate": "0/0",
	            "time_base": "1/90000",
	            "start_pts": 1004578700,
	            "start_time": "11161.985556",
	            "duration_ts": 898612,
	            "duration": "9.984578",
	            "bit_rate": "132792",
	            "disposition": {
	                "default": 0,
	                "dub": 0,
	                "original": 0,
	                "comment": 0,
	                "lyrics": 0,
	                "karaoke": 0,
	                "forced": 0,
	                "hearing_impaired": 0,
	                "visual_impaired": 0,
	                "clean_effects": 0,
	                "attached_pic": 0,
	                "timed_thumbnails": 0,
	                "non_diegetic": 0,
	                "captions": 0,
	                "descriptions": 0,
	                "metadata": 0,
	                "dependent": 0,
	                "still_image": 0
	            }
	        }
	    ]
	}
*/

type Fraction struct {
	Value    string  `json:"value"`
	Dividend int     `json:"dividend"`
	Divisor  int     `json:"divisor"`
	Quotient float64 `json:"quotient"`
}

func (v *Fraction) RoundedQuotient() int {
	return int(math.Round(v.Quotient))
}

func (v *Fraction) UnmarshalJSON(data []byte) error {

	err := json.Unmarshal(data, &v.Value)
	if err != nil {
		return err
	}

	fr := strings.Split(v.Value, "/")
	if len(fr) == 2 {
		v.Dividend, _ = strconv.Atoi(fr[0])
		v.Divisor, _ = strconv.Atoi(fr[1])
		v.Quotient = float64(v.Dividend) / float64(v.Divisor)
	}
	return nil
}

type StreamData struct {
	Index       int    `json:"index,omitempty"`
	CodecType   string `json:"codec_type,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	CodedWidth  int    `json:"coded_width,omitempty"`
	CodedHeight int    `json:"coded_height,omitempty"`

	RFrameRate   Fraction `json:"r_frame_rate,omitempty"`
	AVGFrameRate Fraction `json:"avg_frame_rate,omitempty"`
}

type FormatData struct {
	FileName string `json:"filename,omitempty"`
	Format   string `json:"format,omitempty"`
}

type MetaData struct {
	Streams []StreamData `json:"streams,omitempty"`
	Format  FormatData   `json:"format,omitempty"`
}

func (m *MetaData) GetDimension() (int, int) {
	for _, stream := range m.Streams {
		if stream.CodecType == "video" {
			return stream.Width, stream.Height
		}
	}
	return 0, 0
}

func (m *MetaData) GetVideoStream() *StreamData {
	for _, stream := range m.Streams {
		if stream.CodecType == "video" {
			return &stream
		}
	}
	return nil
}

type metaManager struct {
	meta      map[string]*MetaData
	metaMutex sync.Mutex

	pendingLoad  []string
	pendingMutex sync.Mutex
}

func (c *metaManager) getMeta(remoteId string) *MetaData {
	c.metaMutex.Lock()
	defer c.metaMutex.Unlock()

	if c.meta == nil {
		return nil
	}
	return c.meta[remoteId]
}

func (c *metaManager) addMeta(remoteId string, data *MetaData) {
	if data == nil {
		return
	}
	if remoteId == "" {
		return
	}

	c.metaMutex.Lock()
	defer c.metaMutex.Unlock()

	if c.meta == nil {
		c.meta = map[string]*MetaData{}
	}
	c.meta[remoteId] = data
}

func (c *metaManager) removePending(channelRemoteId string) {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()

	c.pendingLoad = utils.Remove(c.pendingLoad, channelRemoteId, false)
}

func (c *metaManager) isPending(channelRemoteId string) bool {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()

	return utils.Contains(c.pendingLoad, channelRemoteId)
}

func (c *metaManager) waitPending(channelRemoteId string) {
	c.pendingMutex.Lock()
	//defer c.pendingMutex.Unlock()

	for utils.Contains(c.pendingLoad, channelRemoteId) {
		c.pendingMutex.Unlock()
		time.Sleep(time.Millisecond * time.Duration(100))
		c.pendingMutex.Lock()
	}
	c.pendingLoad = utils.AddIfNotExist(c.pendingLoad, channelRemoteId)
	c.pendingMutex.Unlock()
}

var channelMeta metaManager

func LoadMetaData(channelRemoteId string, url string) *MetaData {

	channelMeta.waitPending(channelRemoteId)
	defer channelMeta.removePending(channelRemoteId)

	m := channelMeta.getMeta(channelRemoteId)
	if m != nil {
		return m
	}

	log.Println("Loading:", url)
	cmd := exec.Command("ffprobe", "-timeout", "20", "-v", "quiet", "-print_format", "json", "-show_streams", "-i", url)

	// Use a bytes.Buffer to get the output
	var buf bytes.Buffer
	cmd.Stdout = &buf

	//log.Printf("Start cmd: %s", strings.Join(cmd.Args, " "))
	err := cmd.Start()
	if err != nil {
		log.Printf("Failed to get file info: %v", err)
		return nil
	}

	//log.Printf("Wait cmd")
	err = cmd.Wait()
	if err != nil {
		log.Printf("Failed to wait till cmd finish: %v", err)
		return nil
	}

	var metaData MetaData
	err = json.Unmarshal(buf.Bytes(), &metaData)

	if err != nil {
		log.Printf("Failed Unmarshall metadata: %v", err)
		return nil
	}

	vidStream := metaData.GetVideoStream()
	if vidStream != nil && vidStream.Width != 0 && vidStream.Height != 0 {
		channelMeta.addMeta(channelRemoteId, &metaData)
	}

	return &metaData
}
