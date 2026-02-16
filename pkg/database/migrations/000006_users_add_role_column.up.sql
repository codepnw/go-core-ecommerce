CREATE TYPE user_role AS ENUM ('admin', 'user');

ALTER TABLE users ADD COLUMN role user_role DEFAULT 'user';

-- password = adminexample
INSERT INTO users (email, password, role)
VALUES ('admin@example.com', '$2a$12$gvMU8QokgN1bMVXK0n2OPe1RCG4xfxPs.jK5FygkoN6vM8CrweBiy', 'admin');