package db

import (
	"errors"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

type Channel struct {
	Id          int
	Name        string
	RemoteId    string
	Width       int
	Height      int
	HistoryDays int
	Group       string
}

func QueryGetChannelInfo(remoteId string) (*Channel, error) {

	if remoteId == "" {
		return nil, errors.New("zero remoteId")
	}

	row, err := QueryRow(`SELECT c.id, c.name, c.remote_id, c.width, c.height, c.history_days, c.group_origin
from channel c where c.remote_id = $1;`, remoteId)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if row == nil {
		return nil, errors.New("failed to fetch channel")
	}

	channel := Channel{}
	err = ScanRow(row, &channel.Id, &channel.Name, &channel.RemoteId, &channel.Width, &channel.Height,
		&channel.HistoryDays, &channel.Group)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &channel, err
}

func QueryAddChannelInfo(channel *Channel) error {
	if channel == nil {
		return errors.New("empty channel data")
	}

	row, err := QueryRow(`insert into channel(name, remote_id, width, height, history_days, group_origin)
values($1, $2, $3, $4, $5, $6)
returning id;`, channel.Name, channel.RemoteId, channel.Width, channel.Height, channel.HistoryDays, channel.Group)

	if err != nil {
		log.Println(err)
	}

	if row == nil {
		if err == nil {
			return errors.New("failed insert follow up data to DB")
		}
		return err
	}

	return ScanRow(row, &channel.Id)
}

func QueryUpdateChannel(channel *Channel) error {
	if channel == nil {
		return errors.New("empty channel data")
	}

	count, err := Exec(`update channel set 
name = $2,
history_days = $3,
width = $4,
height = $5,
group_origin = $6,
updated_at = now()
where remote_id = $1
returning id;`, channel.RemoteId, channel.Name, channel.HistoryDays, channel.Width, channel.Height, channel.Group)

	if err != nil {
		log.Println(err)
	}

	if err == nil && count == 0 {
		err = errors.New("zero rows updated")
	}

	return err
}
