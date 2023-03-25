package xmltv

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"m3u8/db"
	"m3u8/util"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

const xmldateformat = "20060102150400 -0700"

type XmlTv struct {
	XMLName       xml.Name       `xml:"tv"`
	ChannelList   []XmlChannel   `xml:"channel"`
	ProgrammeList []XmlProgramme `xml:"programme"`
}

type XmlChannel struct {
	Id   string   `xml:"id,attr"`
	Name []string `xml:"display-name"`
	Icon Icon     `xml:"icon"`
}

type Icon struct {
	Src string `xml:"src,attr"`
}

type XmlProgramme struct {
	Start       string `xml:"start,attr"`
	Stop        string `xml:"stop,attr"`
	Channel     string `xml:"channel,attr"`
	Title       string `xml:"title"`
	Description string `xml:"desc"`

	//SubTitle    string   `xml:"sub-title"`
	//Credits     string   `xml:"credits"`
	//Date        string   `xml:"date"`
	//Categories  []string `xml:"category"`
	//Rating      string   `xml:"rating>value"`

	start time.Time
	stop  time.Time
}

func (x *XmlProgramme) Init() {
	x.start, _ = time.Parse(xmldateformat, x.Start)
	x.stop, _ = time.Parse(xmldateformat, x.Stop)

	// convert to local time zone?
	// x.Start = x.start.Format(xmldateformat)
	// x.Stop = x.start.Format(xmldateformat)
}

type TvgChannel struct {
	dbChannel *db.TvgChannel
	Channel   *XmlChannel
	Programme []XmlProgramme
}

func (x *TvgChannel) filterProgramme() {
	if len(x.Programme) < 2 {
		return
	}

	t := time.Now()
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)

	from := midnight.AddDate(0, 0, -x.dbChannel.HistoryDays)
	to := midnight.AddDate(0, 0, 2)

	for i := len(x.Programme) - 1; i >= 0; i-- {
		if x.Programme[i].start.Unix() >= from.Unix() {
			if x.Programme[i].start.Unix() <= to.Unix() {
				continue
			}
		}
		x.Programme = util.Remove(x.Programme, i)
	}
	sort.SliceStable(x.Programme, func(i, j int) bool {
		return x.Programme[i].start.Unix() < x.Programme[j].start.Unix()
	})

}

func DownloadFullTvGuide(url string, fileName string) (int64, error) {

	err := os.MkdirAll(path.Dir(fileName), 0750)
	if err != nil {
		return 0, err
	}

	out, err := os.Create(fileName)
	if out != nil {
		defer out.Close()
	}
	if err != nil {
		return 0, err
	}

	client := http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return 0, err
	}
	return io.Copy(out, resp.Body)
}

func findChannelById(channels []*TvgChannel, xmlChnlId string) *TvgChannel {
	if channels == nil {
		return nil
	}
	for _, channel := range channels {
		if channel.Channel != nil && channel.Channel.Id == xmlChnlId {
			return channel
		}
	}
	return nil
}

func findChannel(channels []*TvgChannel, channelNames []string) *TvgChannel {
	if channels == nil {
		return nil
	}
	for _, channel := range channels {
		for _, name := range channelNames {
			if channel.dbChannel.TvgName == name {
				return channel
			}
		}
	}
	return nil
}

func GenerateTvGuideFromUrl(conf map[string]string) error {

	url := conf["input_url"]
	inFileName := conf["input_path"]
	outputName := conf["epg_path"]

	log.Printf("Downloading EPG %s", url)
	n, err := DownloadFullTvGuide(url, inFileName)
	if err != nil {
		return err
	}
	log.Printf("Loaded %d bytes", n)

	if strings.HasSuffix(inFileName, ".tar.gz") {
		log.Printf("Extracting tar.gz")
		err = ExtractTarGz(inFileName)
		if err != nil {
			log.Printf("ExtractTarGz error: %v", err)
		}
	} else if strings.HasSuffix(inFileName, ".gz") {
		log.Printf("Extracting gz")

		newfilename := strings.TrimSuffix(inFileName, ".gz")
		err = ExtractGz(inFileName, newfilename)
		if err != nil {
			log.Printf("ExtractTarGz error: %v", err)
		}
		inFileName = newfilename
	}

	return GenerateTvGuide(inFileName, outputName)
}

func GenerateTvGuide(fileName string, outputName string) error {
	tvg, err := extractTvGuide(fileName)
	if err != nil {
		return err
	}

	// Verify that there is used only single name for channel
	for i := len(tvg) - 1; i >= 0; i-- {
		for j := i - 1; j >= 0; j-- {
			if tvg[i].dbChannel != tvg[j].dbChannel {
				if tvg[i].Channel == tvg[j].Channel {
					// TODO: Merge later?
					log.Printf("found same tvguide with multiple name usage: %s, %s", tvg[i].dbChannel.TvgName, tvg[j].dbChannel.TvgName)
				}
			}
		}
		tvg[i].Channel.Name = []string{tvg[i].dbChannel.TvgName}
	}

	var xmlTv XmlTv

	for _, channel := range tvg {
		channel.filterProgramme()
		xmlTv.ChannelList = append(xmlTv.ChannelList, *channel.Channel)
		xmlTv.ProgrammeList = append(xmlTv.ProgrammeList, channel.Programme...)
	}

	data, _ := xml.MarshalIndent(xmlTv, "", " ")

	out, err := os.Create(outputName)
	if out != nil {
		defer out.Close()
	}
	if err != nil {
		return err
	}
	_, err = out.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	_, err = io.Copy(out, bytes.NewReader(data))
	return err
}

func extractTvGuide(fileName string) ([]*TvgChannel, error) {

	tvg, err := db.QueryGetTvgArray()
	if err != nil {
		return nil, fmt.Errorf("QueryGetTvgArray error: %v", err)
	}
	if len(tvg) == 0 {
		return nil, errors.New("empty tvg array")
	}

	outTvg := make([]*TvgChannel, 0, len(tvg))
	for _, channel := range tvg {
		outTvg = append(outTvg, &TvgChannel{
			dbChannel: channel,
		})
	}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("open file error: %v", err)
	}

	decoder := xml.NewDecoder(f)

	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			// If we just read a StartElement token
			// ...and its name is "tv"
			if se.Name.Local == "tv" {

			} else if se.Name.Local == "channel" {
				var c XmlChannel
				err = decoder.DecodeElement(&c, &se)
				if err != nil {
					return nil, fmt.Errorf("failed DecodeElement channel")
				}
				chnl := findChannel(outTvg, c.Name)
				if chnl != nil {
					chnl.Channel = &c
				}

				// TODO: Collect tvguide DB?
				//chnls := strings.Join(c.Name, ", ")
				//log.Printf(chnls)

			} else if se.Name.Local == "programme" {
				var p XmlProgramme
				err = decoder.DecodeElement(&p, &se)
				p.Init()
				if err != nil {
					return nil, fmt.Errorf("failed DecodeElement programme")
				}
				ch := findChannelById(outTvg, p.Channel)
				if ch != nil {
					ch.Programme = append(ch.Programme, p)
				}
			}
		}
	}
	return outTvg, nil
}

func ExtractTarGz(outFileName string) error {
	gzipStream, err := os.Open(outFileName)

	if gzipStream != nil {
		defer gzipStream.Close()
	}
	if err != nil {
		return fmt.Errorf("ExtractTarGz: File open failed %v", err)
	}

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if uncompressedStream != nil {
		defer uncompressedStream.Close()
	}
	if err != nil {
		return fmt.Errorf("ExtractTarGz: NewReader failed %v", err)
	}

	tarReader := tar.NewReader(uncompressedStream)

	for true {
		var header *tar.Header
		header, err = tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("ExtractTarGz: Next() failed: %v", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.Mkdir(header.Name, 0755); err != nil {
				return fmt.Errorf("ExtractTarGz: Mkdir() failed: %v", err)
			}
		case tar.TypeReg:
			var outFile *os.File
			outFile, err = os.Create(header.Name)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err = io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			return fmt.Errorf("ExtractTarGz: uknown type: %s in %s", header.Typeflag, header.Name)
		}

	}
	return nil
}

func ExtractGz(inFileName string, outFileName string) error {

	if inFileName == "" {
		return errors.New("invalid file name")
	}

	gzipfile, err := os.Open(inFileName)

	if err != nil {
		return fmt.Errorf("file open error %v")
	}

	reader, err := gzip.NewReader(gzipfile)

	if err != nil {
		return fmt.Errorf("gzip.NewReader error %v")
	}
	defer reader.Close()

	writer, err := os.Create(outFileName)

	if err != nil {
		return fmt.Errorf("output file create error %v")
	}

	defer writer.Close()

	_, err = io.Copy(writer, reader)
	return err
}
