package db

import (
	"errors"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"time"
)

type Channel struct {
	Id       int64
	TvgName  string
	RemoteId string

	Width     int
	Height    int
	FrameRate int

	CreatedAt time.Time
	UpdatedAt time.Time

	ChannelName ChannelName
}

func (c *Channel) HasAllMeta() bool {
	return c.Width != 0 && c.Height != 0 && c.FrameRate != 0
}

type ChannelName struct {
	Id          int64
	Name        string
	Group       string
	HistoryDays int

	Provider Provider

	CreatedAt time.Time
	UpdatedAt time.Time
}

type TvgChannel struct {
	TvgName     string
	HistoryDays int
}

func QueryInsertOrUpdateChannel(channel *Channel) error {

	if channel == nil {
		return errors.New("empty channel data")
	}

	row, err := QueryRow(`with existing_channel AS (
    select ec.* from channel ec
    where ec.remote_id = $1
), updated_channel AS (
UPDATE channel c_new set width = $2, height = $3, frame_rate = $4,
    updated_at=case when (c_new.width = $2 and c_new.height = $3 and c_new.frame_rate = $4) then c_new.updated_at else now() end
    from existing_channel c_old
    WHERE c_new.id = c_old.id
    returning c_new.id, c_new.created_at, c_new.updated_at, to_jsonb(c_old) as old, to_jsonb(c_new) as new),
inserted_channel AS (
 INSERT INTO channel(remote_id, width, height, frame_rate)
     SELECT $1, $2, $3, $4
     WHERE NOT EXISTS (SELECT uuc.id FROM updated_channel uuc)
     returning id, created_at, updated_at, null::jsonb as old, to_jsonb(channel) as new)
SELECT ic.id, ic.created_at, ic.updated_at, ic.old, ic.new
FROM   inserted_channel ic
UNION  ALL
SELECT uc.id, uc.created_at, uc.updated_at, uc.old, uc.new
FROM updated_channel uc
limit 1;`, channel.RemoteId, channel.Width, channel.Height, channel.FrameRate)

	if err != nil {
		log.Println(err)
	}

	if row == nil {
		if err == nil {
			return errors.New("failed insert follow up data to DB")
		}
		return err
	}

	oldJson := map[string]interface{}{}
	newJson := map[string]interface{}{}

	err = ScanRow(row, &channel.Id, &channel.CreatedAt, &channel.UpdatedAt, &oldJson, &newJson)

	if err != nil {
		return err
	}

	if channel.Id == 0 {
		return errors.New("failed insert or update channel")
	}

	if len(oldJson) != 0 {
		go QueryAddHistory("channel", channel.Id, oldJson, newJson)
	}

	return QueryAddOrUpdateChannelName(channel.Id, &channel.ChannelName)
}

func QueryAddOrUpdateChannelName(channelId int64, channelName *ChannelName) error {
	if channelName == nil {
		return errors.New("empty channel name data")
	}

	err := QueryInsertOrUpdateProvider(&channelName.Provider)

	if err != nil {
		return err
	}

	if channelName.Provider.Id == 0 {
		return errors.New("failed to update providers data")
	}

	row, err := QueryRow(`with existing_channel_n AS (
    select ecn.* from channel_name ecn
    WHERE ecn.channel_id = $1 and ecn.provider_id = $2
), updated_channel_n AS (
UPDATE channel_name cn_new set name = $3, history_days = $4, group_origin = $5,
    updated_at=case when (cn_new.name = $3 and cn_new.history_days = $4 and cn_new.group_origin = $5) then cn_new.updated_at else now() end
    from existing_channel_n cn_old
    WHERE cn_new.id = cn_old.id
    returning cn_new.id, cn_new.created_at, cn_new.updated_at, to_jsonb(cn_old) as old, to_jsonb(cn_new) as new),
inserted_channel_n AS (
 INSERT INTO channel_name(channel_id, provider_id, name, history_days, group_origin)
     SELECT $1, $2, $3, $4, $5
     WHERE NOT EXISTS (SELECT ucn.id FROM updated_channel_n ucn)
     returning id, created_at, updated_at, null::jsonb as old, to_jsonb(channel_name) as new)
SELECT icn.id, icn.created_at, icn.updated_at, icn.old, icn.new
FROM   inserted_channel_n icn
UNION  ALL
SELECT ucn.id, ucn.created_at, ucn.updated_at, ucn.old, ucn.new
FROM updated_channel_n ucn
limit 1;`, channelId, channelName.Provider.Id, channelName.Name, channelName.HistoryDays, channelName.Group)

	if row == nil {
		if err == nil {
			return errors.New("failed insert or update channel name")
		}
		return err
	}

	oldJson := map[string]interface{}{}
	newJson := map[string]interface{}{}

	err = ScanRow(row, &channelName.Id, &channelName.CreatedAt, &channelName.UpdatedAt, &oldJson, &newJson)

	if err != nil {
		return err
	}

	if channelName.Id == 0 {
		return errors.New("failed insert or update channel_name")
	}

	if len(oldJson) != 0 {
		go QueryAddHistory("channel_name", channelName.Id, oldJson, newJson)
	}

	return err
}

func QueryGetChannelInfo(remoteId string, provider *Provider) (*Channel, error) {

	if remoteId == "" {
		return nil, errors.New("zero remoteId")
	}
	if provider == nil {
		return nil, errors.New("invalid provider")
	}

	row, err := QueryRow(`SELECT c.id, c.width, c.height, c.frame_rate, c.created_at, c.updated_at, c.tvg_name,
cn.id, cn.name, cn.history_days, cn.group_origin, cn.created_at, cn.updated_at, p.id, p.name
from channel c
left join providers p on p.host = $2
left join channel_name cn on c.id = cn.channel_id and cn.provider_id = p.id
where c.remote_id = $1
order by c.id, cn.updated_at DESC NULLS LAST limit 1;`, remoteId, provider.Host)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if row == nil {
		return nil, errors.New("failed to fetch channel")
	}

	channel := Channel{RemoteId: remoteId, ChannelName: ChannelName{Provider: *provider}}
	err = ScanRow(row, &channel.Id, &channel.Width, &channel.Height, &channel.FrameRate, &channel.CreatedAt, &channel.UpdatedAt, &channel.TvgName,
		&channel.ChannelName.Id, &channel.ChannelName.Name, &channel.ChannelName.HistoryDays, &channel.ChannelName.Group, &channel.ChannelName.CreatedAt, &channel.ChannelName.UpdatedAt,
		&channel.ChannelName.Provider.Id, &channel.ChannelName.Provider.Name)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &channel, err
}

func QueryGetTvgArray() ([]*TvgChannel, error) {

	tvgChannels := make([]*TvgChannel, 0, 10)

	rows, err := QueryRows(`select c.tvg_name, array_agg(DISTINCT cn.history_days)
from channel c
left join channel_name cn on c.id = cn.channel_id
where c.tvg_name is not null and c.tvg_name != '' and c.tvg_generate = true
group by c.tvg_name`)

	if err != nil {
		return nil, err
	}

	if rows == nil {
		return tvgChannels, errors.New("failed to fetch tvgChannels from DB")
	}

	defer rows.Close()

	for rows.Next() {

		tvgChannel := TvgChannel{}

		var hDays []int
		err = ScanRows(rows, &tvgChannel.TvgName, &hDays)

		for _, day := range hDays {
			if day > tvgChannel.HistoryDays {
				tvgChannel.HistoryDays = day
			}
		}

		if err != nil {
			return nil, err
		}
		tvgChannels = append(tvgChannels, &tvgChannel)
	}
	return tvgChannels, err
}
