-- Create table for storing users --
CREATE TABLE IF NOT EXISTS users (
  id SERIAL NOT NULL PRIMARY KEY,
  biometric_id DOUBLE PRECISION[] NOT NULL,
  email TEXT UNIQUE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create function for keeping `updated_at` updated with current timestamp
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
    RETURNS trigger AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$
    LANGUAGE 'plpgsql';

-- Create trigger to call the function we created to keep updated_at updated
DROP trigger IF EXISTS users_updated_at ON users;
CREATE trigger users_updated_at
    BEFORE UPDATE ON users FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();
