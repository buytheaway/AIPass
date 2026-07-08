update users
set password_hash = '$2a$10$7EqJtq98hPqEX7fNZaFWoOHi1PrmGbUQb4nFQWDX2XrtmFgR7VX3W',
    updated_at = now()
where email = 'admin@aipass.local'
  and role = 'admin';

