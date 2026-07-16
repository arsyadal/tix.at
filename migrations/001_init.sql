create table if not exists events (
  id text primary key,
  name text not null,
  stock integer not null check (stock >= 0),
  created_at timestamptz not null default now()
);

create table if not exists bookings (
  id text primary key,
  event_id text not null references events(id),
  user_id text not null,
  status text not null check (status in ('PENDING', 'PAID', 'CANCELLED')),
  expires_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists bookings_event_id_idx on bookings(event_id);
create index if not exists bookings_status_idx on bookings(status);
create index if not exists bookings_expires_at_idx on bookings(expires_at);
