drop table provider_channel;
alter table channel drop column name;
alter table channel drop column group_origin;
alter table channel_name drop column provider;
alter table channel_name add column provider_id  integer;

ALTER TABLE channel ADD CONSTRAINT channel_remote_id_key UNIQUE (remote_id);

create table channel_name
(
    id           bigserial
        primary key,
    channel_id   bigint                                 not null,
    name         text,
    group_origin text,
    created_at   timestamp with time zone default now() not null,
    updated_at   timestamp with time zone,
    provider_id  integer,
    constraint channel_name_pk
        unique (channel_id, provider_id)
);

create table providers
(
    id   serial
        primary key,
    name text,
    host text not null
        constraint providers_pk
            unique
);
