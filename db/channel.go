package db

import (
	"errors"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"time"
)

type Channel struct {
	Id          int
	RemoteId    string
	Width       int
	Height      int
	HistoryDays int

	CreatedAt time.Time
	UpdatedAt time.Time

	ChannelName ChannelName
}

type ChannelName struct {
	Id    int64
	Name  string
	Group string

	Provider Provider

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Provider struct {
	Id   int32
	Name string
	Host string
}

func QueryInsertOrUpdateChannel(channel *Channel) error {

	if channel == nil {
		return errors.New("empty channel data")
	}

	row, err := QueryRow(`insert into channel(remote_id, width, height, history_days)
	values($1, $2, $3, $4)
	on conflict(remote_id) do update
	set width = $2,
		height = $3,
		history_days = $4,
		updated_at=case when (channel.width = $2 and channel.height = $3 and channel.history_days = $4) then channel.updated_at else now() end
	returning id, created_at, updated_at;`, channel.RemoteId, channel.Width, channel.Height, channel.HistoryDays)

	if err != nil {
		log.Println(err)
	}

	if row == nil {
		if err == nil {
			return errors.New("failed insert follow up data to DB")
		}
		return err
	}

	err = ScanRow(row, &channel.Id, &channel.CreatedAt, &channel.UpdatedAt)

	if err != nil {
		return err
	}

	if channel.Id == 0 {
		return errors.New("failed insert or update channel")
	}

	return QueryAddOrUpdateChannelName(channel.Id, &channel.ChannelName)
}

func QueryAddOrUpdateChannelName(channelId int, channel *ChannelName) error {
	if channel == nil {
		return errors.New("empty channel data")
	}

	row, err := QueryRow(`with updated_provider AS (
insert into providers(host)
values($2)
on conflict(host) do update set name=providers.name 
returning id)
insert into channel_name(channel_id, provider_id, name, group_origin)
select $1, up.id, $3, $4
from updated_provider up
on conflict(channel_id, provider_id) do update
set name = $3, group_origin=$4, 
    updated_at=case when (channel_name.name = $3 and channel_name.group_origin = $4) then channel_name.updated_at else now() end
returning id, created_at, updated_at;`, channelId, channel.Provider.Host, channel.Name, channel.Group)

	if row == nil {
		if err == nil {
			return errors.New("failed insert or update channel name")
		}
		return err
	}

	return ScanRow(row, &channel.Id, &channel.CreatedAt, &channel.UpdatedAt)
}

func QueryGetChannelInfo(remoteId string, providerHost string) (*Channel, error) {

	if remoteId == "" {
		return nil, errors.New("zero remoteId")
	}
	if providerHost == "" {
		return nil, errors.New("invalid provider")
	}

	row, err := QueryRow(`SELECT c.id, c.width, c.height, c.history_days, c.created_at, c.updated_at,
cn.id, cn.name, cn.group_origin, cn.created_at, cn.updated_at, p.id, p.host
from channel c
left join providers p on p.host = $2
left join channel_name cn on c.id = cn.channel_id and cn.provider_id = p.id
where c.remote_id = $1;`, remoteId, providerHost)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if row == nil {
		return nil, errors.New("failed to fetch channel")
	}

	channel := Channel{RemoteId: remoteId, ChannelName: ChannelName{Provider: Provider{Host: providerHost}}}
	err = ScanRow(row, &channel.Id, &channel.Width, &channel.Height, &channel.HistoryDays, &channel.CreatedAt, &channel.UpdatedAt,
		&channel.ChannelName.Id, &channel.ChannelName.Name, &channel.ChannelName.Group, &channel.ChannelName.CreatedAt, &channel.ChannelName.UpdatedAt,
		&channel.ChannelName.Provider.Id, &channel.ChannelName.Provider.Name)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &channel, err
}

/*
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
	// TODO: Add or update channel info

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
*/
