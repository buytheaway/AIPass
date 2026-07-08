create extension if not exists pgcrypto;

create table files (
    id uuid primary key,
    bucket text not null,
    object_key text not null,
    original_name text not null,
    content_type text not null,
    size_bytes bigint not null check (size_bytes >= 0),
    created_at timestamptz not null default now(),
    unique (bucket, object_key)
);

create table users (
    id uuid primary key,
    email text unique not null,
    phone text,
    full_name text not null,
    role text not null check (role in ('admin', 'member')),
    password_hash text,
    photo_file_id uuid references files(id),
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table subscription_plans (
    id uuid primary key,
    name text not null,
    description text,
    duration_days int not null check (duration_days > 0),
    price numeric(12,2) not null check (price >= 0),
    currency text not null default 'KZT',
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table user_subscriptions (
    id uuid primary key,
    user_id uuid not null references users(id),
    plan_id uuid not null references subscription_plans(id),
    starts_at timestamptz not null,
    ends_at timestamptz not null,
    status text not null check (status in ('pending_payment', 'active', 'expired', 'cancelled')),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    check (ends_at > starts_at)
);

create table qr_passes (
    id uuid primary key,
    user_id uuid not null references users(id),
    subscription_id uuid not null references user_subscriptions(id),
    token_hash text unique not null,
    status text not null check (status in ('active', 'revoked', 'expired')),
    expires_at timestamptz not null,
    created_at timestamptz not null default now()
);

create table access_events (
    id uuid primary key,
    user_id uuid not null references users(id),
    subscription_id uuid references user_subscriptions(id),
    qr_pass_id uuid references qr_passes(id),
    event_type text not null check (event_type in ('check_in', 'check_out', 'denied')),
    decision text not null check (decision in ('allowed', 'denied')),
    reason text,
    scanner_id text,
    photo_file_id uuid references files(id),
    occurred_at timestamptz not null,
    created_at timestamptz not null default now()
);

create table payments (
    id uuid primary key,
    user_id uuid not null references users(id),
    subscription_id uuid not null references user_subscriptions(id),
    amount numeric(12,2) not null check (amount >= 0),
    currency text not null default 'KZT',
    method text not null check (method in ('kaspi_manual', 'cash', 'bank_transfer')),
    status text not null check (status in ('uploaded', 'approved', 'rejected')),
    receipt_file_id uuid references files(id),
    approved_by uuid references users(id),
    approved_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_users_email on users(email);
create index idx_subscriptions_user_id on user_subscriptions(user_id);
create index idx_qr_passes_token_hash on qr_passes(token_hash);
create index idx_access_events_user_time on access_events(user_id, occurred_at desc);
create index idx_payments_status on payments(status);

insert into users (id, email, full_name, role, password_hash, is_active, created_at, updated_at)
values (
    '00000000-0000-0000-0000-000000000001',
    'admin@aipass.local',
    'Local Admin',
    'admin',
    '$2a$10$Qa29wh5.a6an4ithu/sxme/9IdIBld4fpbl5.tRIYmlp93IyDVX3i',
    true,
    now(),
    now()
) on conflict (email) do nothing;
