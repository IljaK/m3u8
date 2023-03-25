alter table channel_name add column history_days integer default 0;
alter table channel drop column history_days;