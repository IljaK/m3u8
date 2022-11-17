package db

import (
	"errors"
	"log"
)

func QueryUpdateProvider(channel *Channel, provider string) error {
	if channel == nil {
		return errors.New("empty channel data")
	}

	_, err := Exec(`INSERT INTO provider_channel
(provider_type, channel_id) SELECT $2, $1
WHERE NOT EXISTS (SELECT pc.id FROM provider_channel pc WHERE channel_id = $1 and pc.provider_type = $2);`, channel.Id, provider)

	if err != nil {
		log.Println(err)
	}

	if err != nil {
		err = errors.New("zero rows updated")
	}

	return err
}
