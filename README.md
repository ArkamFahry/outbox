# Outbox

Outbox is a side car service which reads the a outbox table from PostgreSQL and pushes those outbox events to NATS

## Outbox table structure

```sql
create or replace function on_events_create()
    returns trigger as
$$
begin
    new.id = 'event_' || gen_random_ulid(); -- gen_random_uuid() or gen_random_ulid()
    new.version = 0;
    new.status = 'pending';
    new.published_at = null;
    new.failed_at = null;
    new.failed_reasons = null;
    new.failed_attempts = 0;
    new.retried_at = null;
    new.retried_attempts = 0;
    new.created_at = now();

    return new;
end;
$$ language plpgsql;

create table if not exists events
(
    id               text          not null,
    version          int default 0 not null,
    aggregate_type   text          not null,
    event_type       text          not null,
    payload          jsonb         not null,
    status           text          not null,
    published_at     timestamptz   null,
    failed_at        timestamptz   null,
    failed_reasons   text[]        null,
    failed_attempts  int default 0 not null,
    retried_at       timestamptz   null,
    retried_attempts int default 0 not null,
    created_at       timestamptz   not null,
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
```
