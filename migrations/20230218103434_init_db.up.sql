create table channel
(
    id           serial,
    remote_id    text                                   not null,
    width        integer                                not null,
    height       integer                                not null,
    name         text,
    created_at   timestamp with time zone default now() not null,
    history_days integer                  default 0     not null,
    group_origin text,
    updated_at   timestamp with time zone
);

create unique index channel_id_uindex
    on channel (id);

create table provider_channel
(
    id            serial
        constraint provider_channel_pk
            primary key,
    channel_id    integer not null,
    provider_type text
);

create unique index provider_channel_id_uindex
    on provider_channel (id);