BEGIN;

CREATE DOMAIN uint8 AS smallint 
    CHECK(VALUE >= 0 AND VALUE <= 255);

CREATE DOMAIN uint16 AS integer 
    CHECK(VALUE >= 0 AND VALUE <= 65536);

CREATE TABLE users (
   id serial NOT NULL PRIMARY KEY,
   username text UNIQUE NOT NULL,
   password text NOT NULL
);

CREATE TABLE characters (
  id serial NOT NULL PRIMARY KEY,
  user_id bigint REFERENCES users(id),
  is_female boolean,
  is_new_character boolean,
  small_gr_level uint8,
  gr_override_mode boolean,
  name varchar(15),
  unk_desc_string varchar(31),
  gr_override_level uint16,
  gr_override_unk0 uint8,
  gr_override_unk1 uint8
);

CREATE TABLE sign_sessions (
  id serial NOT NULL PRIMARY KEY,
  user_id bigint REFERENCES users(id),
  auth_token_num bigint,
  auth_token_str text
);

END;