package ffprobe

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os/exec"
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

func LoadMetaData(url string) *MetaData {

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

	return &metaData
}
