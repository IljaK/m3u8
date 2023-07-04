
create table providers
(
    id serial primary key,
    name text,
    host text not null unique
);

create table channel_name
(
    id           bigserial primary key,
    channel_id   bigint not null,
    name         text,
    group_origin text,
    created_at   timestamp with time zone default now() not null,
    updated_at   timestamp with time zone,
    provider_id  integer,
    history_days integer  default 0,
    tvg_name     text
);

alter table channel_name
    add constraint channel_name_pk unique (channel_id, provider_id);

create table channel
(
    id           serial,
    remote_id    text                                   not null unique,
    width        integer                                not null,
    height       integer                                not null,
    created_at   timestamp with time zone default now() not null,
    updated_at   timestamp with time zone,
    tvg_name     text,
    tvg_generate boolean                  default false
);