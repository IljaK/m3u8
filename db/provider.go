package db

import (
	"errors"
	log "github.com/sirupsen/logrus"
)

type Provider struct {
	Id   int32
	Name string
	Host string
}

func QueryInsertOrUpdateProvider(providerHost string, providerName string) (*Provider, error) {
	if providerHost == "" {
		return nil, errors.New("empty provider host")
	}

	row, err := QueryRow(`with existing_provider AS (
    SELECT id, host, name FROM providers WHERE host = $1),
inserted_provider AS (
INSERT INTO providers(host, name)
SELECT $1, $2
WHERE NOT EXISTS (SELECT id FROM existing_provider)
on conflict(host) do update set name = case when providers.name is not null then providers.name else $2 end
returning id, host, name)
SELECT ip.id, ip.host, ip.name
FROM   inserted_provider ip
UNION  ALL
SELECT p.id, p.host, p.name
FROM existing_provider p;`, providerHost, providerName)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	if row == nil {
		if err == nil {
			return nil, errors.New("failed to insert/update provider data")
		}
		return nil, err
	}

	p := Provider{}

	err = ScanRow(row, &p.Id, &p.Host, &p.Name)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &p, err
}
