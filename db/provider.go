package db

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"strings"
)

type Provider struct {
	Id   int32
	Name string
	Host string

	SubDomain string
	AccessKey string
}

func (p *Provider) FromUri(host string, path []string) {

	// http://wkejhfk.rossteleccom.net/iptv/ABCD3HG7DW38ZD/205/index.m3u8
	// host + / + "iptv" + / + key + / + channel_id + / + file

	args := strings.Split(host, ".")
	p.SubDomain = args[0]
	if len(args) > 1 {
		p.Host = strings.Join(args[1:], ".")
	} else {
		p.Host = host
	}

	if len(path) >= 3 {
		p.AccessKey = path[2]
	}
}

func QueryInsertOrUpdateProvider(provider *Provider) error {
	if provider == nil {
		return errors.New("empty provider data")
	}

	row, err := QueryRow(`with existing_provider AS (
    SELECT id, host, name FROM providers WHERE host = $1),
inserted_provider AS (
INSERT INTO providers(host, name)
SELECT $1, $2
WHERE NOT EXISTS (SELECT id FROM existing_provider)
on conflict(host) do update set name = case when providers.name is not null then providers.name else $2 end
returning id, host, name)
SELECT ip.id
FROM   inserted_provider ip
UNION  ALL
SELECT p.id
FROM existing_provider p;`, provider.Host, provider.Name)

	if err != nil {
		log.Error(err)
		return err
	}

	if row == nil {
		if err == nil {
			return errors.New("failed to insert/update provider data")
		}
		return err
	}

	err = ScanRow(row, &provider.Id)

	if err != nil {
		log.Error(err)
	}

	return err
}
