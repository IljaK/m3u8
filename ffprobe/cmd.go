package ffprobe

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	utils "m3u8/util"
	"os/exec"
	"sync"
	"time"
)

type StreamData struct {
	Index       int    `json:"index,omitempty"`
	CodecType   string `json:"codec_type,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	CodedWidth  int    `json:"coded_width,omitempty"`
	CodedHeight int    `json:"coded_height,omitempty"`
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

	c.pendingLoad = utils.Remove(c.pendingLoad, channelRemoteId)
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
