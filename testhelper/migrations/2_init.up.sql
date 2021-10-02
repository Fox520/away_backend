CREATE TABLE IF NOT EXISTS subscription_status (
    s_status VARCHAR(255) PRIMARY KEY,
    s_description VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS users(
    id TEXT,
    username TEXT NOT NULL,
    email VARCHAR(255) NOT NULL,
    device_token TEXT,
    bio TEXT,
    verified BOOLEAN,
    s_status VARCHAR(255) NOT NULL,
    createdAt TIMESTAMP without time zone DEFAULT (now() at time zone 'utc'),
    profile_picture_url VARCHAR,
    PRIMARY KEY (id),
    CONSTRAINT fk_subscription_status
      FOREIGN KEY(s_status) 
        REFERENCES subscription_status(s_status)
          ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS subscriptions (
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    s_status VARCHAR(255) NOT NULL,
    CONSTRAINT fk_subscription_status
      FOREIGN KEY(s_status) 
        REFERENCES subscription_status(s_status)
          ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS verification_stage (
    v_stage VARCHAR(255) PRIMARY KEY,
    v_description VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS user_verifications (
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    createdAt TIMESTAMP without time zone DEFAULT (now() at time zone 'utc'),
    file_urls TEXT, -- links separated by delimiter
    v_stage VARCHAR(255),
    CONSTRAINT fk_verification_stage
      FOREIGN KEY(v_stage) 
        REFERENCES verification_stage(v_stage)
          ON UPDATE CASCADE
);


CREATE TABLE IF NOT EXISTS property_usage (
    id smallint  GENERATED ALWAYS AS IDENTITY,
    p_usage text,
    usage_description text,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS property_type (
    id smallint GENERATED ALWAYS AS IDENTITY,
    p_type text,
    type_description text,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS property_category (
    id smallint GENERATED ALWAYS AS IDENTITY,
    p_category text,
    category_description text,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS property_combinations (
  type_id smallint REFERENCES property_type(id) ON DELETE CASCADE,
  category_id smallint REFERENCES property_category(id) ON DELETE CASCADE,
  CONSTRAINT property_combinations_pkey PRIMARY KEY (type_id, category_id)
);

CREATE TABLE IF NOT EXISTS properties (
    id uuid DEFAULT uuid_generate_v4(),
    user_id text NOT NULL,
    property_type_id smallint NOT NULL,
    property_category_id smallint NOT NULL,
    property_usage_id smallint NOT NULL,
    bedrooms smallint NOT NULL,
    bathrooms smallint NOT NULL,
    surburb text,
    town text,
    title text NOT NULL,
    p_description text,
    currency text NOT NULL,
    available boolean DEFAULT true,
    price numeric NOT NULL,
    deposit numeric NOT NULL,
    sharing_price numeric DEFAULT 0.000000,
    promoted boolean DEFAULT false,
    posted_date timestamp without time zone default (now() at time zone 'utc'),
    pets_allowed boolean DEFAULT false,
    free_wifi boolean DEFAULT false,
    water_included boolean,
    electricity_included boolean,
    latitude numeric,
    longitude numeric,
    PRIMARY KEY(id),
    CONSTRAINT fk_user
      FOREIGN KEY(user_id) 
        REFERENCES users(id)
          ON DELETE CASCADE,
    CONSTRAINT fk_property_type_id
      FOREIGN KEY(property_type_id) 
        REFERENCES property_type(id)
          ON DELETE CASCADE,
    CONSTRAINT fk_property_category_id
      FOREIGN KEY(property_category_id) 
        REFERENCES property_category(id)
          ON DELETE CASCADE,
    CONSTRAINT fk_property_usage_id
      FOREIGN KEY(property_usage_id) 
        REFERENCES property_usage(id)
          ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS property_photos (
    id uuid DEFAULT uuid_generate_v4(),
    property_id uuid,
    p_url text,
    PRIMARY KEY (id),
    CONSTRAINT fk_property
      FOREIGN KEY(property_id) 
        REFERENCES properties(id)
          ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS featured_areas (
    id serial,
    title text,
    photo_url text,
    latitude numeric,
    longitude numeric,
    country text,
    PRIMARY KEY (id)
);

-- geo query index

CREATE INDEX IF NOT EXISTS latlngindx
   ON properties USING gist (ll_to_earth(latitude, longitude));

-- trigger
CREATE OR REPLACE FUNCTION create_user_subscription() RETURNS TRIGGER AS $example_table$
   BEGIN
      INSERT INTO subscriptions (user_id, s_status) VALUES (new.ID, new.s_status);
      RETURN NEW;
   END;
$example_table$ LANGUAGE plpgsql;

CREATE TRIGGER user_subscription_create_trigger AFTER INSERT ON users
FOR EACH ROW EXECUTE PROCEDURE create_user_subscription();

-- reports
CREATE TABLE IF NOT EXISTS user_report_reasons (
    r_reason VARCHAR(255) PRIMARY KEY,
    r_description TEXT
);

CREATE TABLE IF NOT EXISTS property_report_reasons (
    r_reason VARCHAR(255) PRIMARY KEY,
    r_description TEXT
);

CREATE TABLE IF NOT EXISTS user_reports (
    reporter_user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    reported_user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    createdAt TIMESTAMP without time zone DEFAULT (now() at time zone 'utc'),
    r_reason VARCHAR(255) NOT NULL,
    CONSTRAINT user_reports_pkey PRIMARY KEY (reporter_user_id, reported_user_id),
    CONSTRAINT fk_reason
      FOREIGN KEY(r_reason) 
        REFERENCES user_report_reasons(r_reason)
          ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS property_reports (
    reporter_user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    reported_property_id uuid REFERENCES properties(id) ON DELETE CASCADE,
    createdAt TIMESTAMP without time zone DEFAULT (now() at time zone 'utc'),
    r_reason VARCHAR(255) NOT NULL,
    CONSTRAINT property_reports_pkey PRIMARY KEY (reporter_user_id, reported_property_id),
    CONSTRAINT fk_reason
      FOREIGN KEY(r_reason) 
        REFERENCES property_report_reasons(r_reason)
          ON DELETE CASCADE
);

-- bookings

CREATE TABLE IF NOT EXISTS bookings (
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    property_id uuid REFERENCES properties(id) ON DELETE CASCADE,
    booking_date TIMESTAMP without time zone NOT NULL,
    additional_info TEXT,
    CONSTRAINT user_product_pkey PRIMARY KEY (user_id, property_id)
);