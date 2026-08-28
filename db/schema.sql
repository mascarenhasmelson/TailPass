
-- CREATE TABLE groups (
--     id SERIAL PRIMARY KEY,
--     name TEXT NOT NULL UNIQUE,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );
DROP TABLE IF EXISTS services;
CREATE TABLE services (
  id SERIAL PRIMARY KEY,
  -- group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  service_name VARCHAR(100) NOT NULL,
  local_ip INET NOT NULL,
  local_port INTEGER NOT NULL CHECK (local_port BETWEEN 1 AND 65535),
  remote_ip INET NOT NULL,
  remote_port INTEGER NOT NULL CHECK (remote_port BETWEEN 1 AND 65535),
  online BOOLEAN DEFAULT FALSE,  
  last_seen TIMESTAMP, 
  pid INTEGER,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  -- UNIQUE(service_name, group_id)
);

-- Single-admin authentication. The backend also creates these tables itself
-- on startup (idempotently) so upgrading an existing TailPass deployment
-- doesn't require wiping this init-only schema's Postgres volume.
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  username VARCHAR(64) NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
