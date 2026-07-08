update users
set password_hash = '$2a$10$Qa29wh5.a6an4ithu/sxme/9IdIBld4fpbl5.tRIYmlp93IyDVX3i',
    updated_at = now()
where email = 'admin@aipass.local'
  and role = 'admin';

