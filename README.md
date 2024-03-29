# Outbox

Outbox is a side car service which reads the a outbox table from Postgres and pushes those outbox events to NATS

## Outbox table structure

```sql
-- generate a random ULID
create extension if not exists pgcrypto;

create or replace function gen_random_ulid()
    returns text
as
$$
declare
    -- Crockford's Base32
encoding  bytea = '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
timestamp bytea = E'\\000\\000\\000\\000\\000\\000';
output    text  = '';
    unix_time bigint;
    ulid      bytea;
BEGIN
    unix_time = (extract(epoch from clock_timestamp()) * 1000)::bigint;
timestamp = set_byte(timestamp, 0, (unix_time >> 40)::bit(8)::integer);
timestamp = set_byte(timestamp, 1, (unix_time >> 32)::bit(8)::integer);
timestamp = set_byte(timestamp, 2, (unix_time >> 24)::bit(8)::integer);
timestamp = set_byte(timestamp, 3, (unix_time >> 16)::bit(8)::integer);
timestamp = set_byte(timestamp, 4, (unix_time >> 8)::bit(8)::integer);
timestamp = set_byte(timestamp, 5, unix_time::bit(8)::integer);

    -- 10 entropy bytes
    ulid = timestamp || gen_random_bytes(10);

    -- Encode the timestamp
output = output || chr(get_byte(encoding, (get_byte(ulid, 0) & 224) >> 5));
output = output || chr(get_byte(encoding, (get_byte(ulid, 0) & 31)));
output = output || chr(get_byte(encoding, (get_byte(ulid, 1) & 248) >> 3));
output = output || chr(get_byte(encoding, ((get_byte(ulid, 1) & 7) << 2) | ((get_byte(ulid, 2) & 192) >> 6)));
output = output || chr(get_byte(encoding, (get_byte(ulid, 2) & 62) >> 1));
output = output || chr(get_byte(encoding, ((get_byte(ulid, 2) & 1) << 4) | ((get_byte(ulid, 3) & 240) >> 4)));
output = output || chr(get_byte(encoding, ((get_byte(ulid, 3) & 15) << 1) | ((get_byte(ulid, 4) & 128) >> 7)));
output = output || chr(get_byte(encoding, (get_byte(ulid, 4) & 124) >> 2));
output = output || chr(get_byte(encoding, ((get_byte(ulid, 4) & 3) << 3) | ((get_byte(ulid, 5) & 224) >> 5)));
output = output || chr(get_byte(encoding, (get_byte(ulid, 5) & 31)));

    -- Encode the entropy
output = output || chr(get_byte(encoding, (get_byte(ulid, 6) & 248) >> 3));
output = output || chr(get_byte(encoding, ((get_byte(ulid, 6) & 7) << 2) | ((get_byte(ulid, 7) & 192) >> 6)));
output = output || chr(get_byte(encoding, (get_byte(ulid, 7) & 62) >> 1));
output = output || chr(get_byte(encoding, ((get_byte(ulid, 7) & 1) << 4) | ((get_byte(ulid, 8) & 240) >> 4)));
output = output || chr(get_byte(encoding, ((get_byte(ulid, 8) & 15) << 1) | ((get_byte(ulid, 9) & 128) >> 7)));
output = output || chr(get_byte(encoding, (get_byte(ulid, 9) & 124) >> 2));
output = output || chr(get_byte(encoding, ((get_byte(ulid, 9) & 3) << 3) | ((get_byte(ulid, 10) & 224) >> 5)));
output = output || chr(get_byte(encoding, (get_byte(ulid, 10) & 31)));
output = output || chr(get_byte(encoding, (get_byte(ulid, 11) & 248) >> 3));
output = output || chr(get_byte(encoding, ((get_byte(ulid, 11) & 7) << 2) | ((get_byte(ulid, 12) & 192) >> 6)));
output = output || chr(get_byte(encoding, (get_byte(ulid, 12) & 62) >> 1));
output = output || chr(get_byte(encoding, ((get_byte(ulid, 12) & 1) << 4) | ((get_byte(ulid, 13) & 240) >> 4)));
output = output || chr(get_byte(encoding, ((get_byte(ulid, 13) & 15) << 1) | ((get_byte(ulid, 14) & 128) >> 7)));
output = output || chr(get_byte(encoding, (get_byte(ulid, 14) & 124) >> 2));
output = output || chr(get_byte(encoding, ((get_byte(ulid, 14) & 3) << 3) | ((get_byte(ulid, 15) & 224) >> 5)));
output = output || chr(get_byte(encoding, (get_byte(ulid, 15) & 31)));

    RETURN output;
END
$$
LANGUAGE plpgsql
    VOLATILE;

create or replace function on_events_create()
    returns trigger as
$$
begin
    new.id = 'event_' || gen_random_ulid(); -- gen_random_uuid() or gen_random_ulid()
    new.version = 0;
    new.status = 'pending';
    new.published_at = null;
    new.created_at = now();
    new.updated_at = null;

    return new;
end;
$$ language plpgsql;

create or replace function on_events_update()
    returns trigger as
$$
begin
    new.version = new.version + 1;
    new.updated_at = now();

return new;
end;
$$ language plpgsql;

create table if not exists events
(
    id               text          not null,
    version          int default 0 not null,
    aggregate_type   text          not null,
    event_type       text          not null,
    content          jsonb         not null,
    status           text          not null,
    published_at     timestamptz   null,
    created_at       timestamptz   not null,
    updated_at       timestamptz   null,
    constraint events_id_primary_key primary key (id),
    constraint events_id_version_unique unique (id, version),
    constraint events_id_check check ( trim(id) <> '' ),
    constraint events_version_check check ( version >= 0 ),
    constraint events_aggregate_type_check check ( trim(aggregate_type) <> '' ),
    constraint events_event_type_check check ( trim(event_type) <> '' ),
    constraint events_status_check check ( status in ('pending', 'published', 'failed') )
);

create or replace trigger on_events_create
    before insert
    on events
    for each row
execute function on_events_create();
       
create or replace trigger on_events_update
    before update
    on events
    for each row
execute function on_events_update();
```
