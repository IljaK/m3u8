
create table update_history
(
    id bigserial primary key,
    changed_at timestamptz default now(),
    table_name text not null,
    row_id bigint not null,
    changed_values jsonb
);

alter sequence channel_id_seq as bigint;
alter table channel alter id type bigint;
alter table channel alter frame_rate type integer;