CREATE EXTENSION IF NOT EXISTS cube;
CREATE EXTENSION IF NOT EXISTS earthdistance;

CREATE TABLE IF NOT EXISTS subscription_status (
    s_status VARCHAR(255) PRIMARY KEY,
    s_description VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS users(
    id TEXT,
    username TEXT NOT NULL,
    email VARCHAR(255) NOT NULL,
    device_token TEXT NOT NULL,
    bio TEXT NOT NULL,
    verified BOOLEAN DEFAULT false NOT NULL,
    s_status VARCHAR(255) NOT NULL,
    profile_picture_url VARCHAR NOT NULL,
    createdAt TIMESTAMP without time zone DEFAULT (now() at time zone 'utc') NOT NULL,
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
    p_usage text NOT NULL,
    usage_description text,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS property_type (
    id smallint GENERATED ALWAYS AS IDENTITY,
    p_type text NOT NULL,
    type_description text,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS property_category (
    id smallint GENERATED ALWAYS AS IDENTITY,
    p_category text NOT NULL,
    category_description text,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS property_combinations (
  type_id smallint REFERENCES property_type(id) ON DELETE CASCADE,
  category_id smallint REFERENCES property_category(id) ON DELETE CASCADE,
  CONSTRAINT property_combinations_pkey PRIMARY KEY (type_id, category_id)
);

CREATE TABLE IF NOT EXISTS properties (
    id uuid DEFAULT gen_random_uuid(),
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
    available boolean DEFAULT true NOT NULL,
    price numeric NOT NULL,
    deposit numeric NOT NULL,
    sharing_price numeric DEFAULT 0.000000,
    promoted boolean DEFAULT false NOT NULL,
    posted_date timestamp without time zone default (now() at time zone 'utc') NOT NULL,
    pets_allowed boolean DEFAULT false NOT NULL,
    free_wifi boolean DEFAULT false NOT NULL,
    water_included boolean DEFAULT false NOT NULL,
    electricity_included boolean DEFAULT false NOT NULL,
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
    id uuid DEFAULT gen_random_uuid(),
    property_id uuid NOT NULL,
    p_url text NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_property
      FOREIGN KEY(property_id) 
        REFERENCES properties(id)
          ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS featured_areas (
    id serial,
    title text NOT NULL,
    photo_url text NOT NULL,
    latitude numeric NOT NULL,
    longitude numeric NOT NULL,
    country text NOT NULL,
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
    r_description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS property_report_reasons (
    r_reason VARCHAR(255) PRIMARY KEY,
    r_description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_reports (
    reporter_user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    reported_user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    createdAt TIMESTAMP without time zone DEFAULT (now() at time zone 'utc') NOT NULL,
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
    createdAt TIMESTAMP without time zone DEFAULT (now() at time zone 'utc') NOT NULL,
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


--
-- PostgreSQL database dump
--

-- Dumped from database version 13.3
-- Dumped by pg_dump version 13.3

INSERT INTO property_category (id, p_category) OVERRIDING SYSTEM VALUE VALUES (1, 'Flat');
INSERT INTO property_category (id, p_category) OVERRIDING SYSTEM VALUE VALUES (2, 'Apartment');
INSERT INTO property_category (id, p_category) OVERRIDING SYSTEM VALUE VALUES (3, 'House');
INSERT INTO property_category (id, p_category) OVERRIDING SYSTEM VALUE VALUES (4, 'Townhouse');
INSERT INTO property_category (id, p_category) OVERRIDING SYSTEM VALUE VALUES (5, 'Villa');
INSERT INTO property_category (id, p_category) OVERRIDING SYSTEM VALUE VALUES (6, 'Inside room');
INSERT INTO property_category (id, p_category) OVERRIDING SYSTEM VALUE VALUES (7, 'Outside room');


--
-- Data for Name: property_type; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO property_type (id, p_type) OVERRIDING SYSTEM VALUE VALUES (1, 'Flat');
INSERT INTO property_type (id, p_type) OVERRIDING SYSTEM VALUE VALUES (2, 'House');
INSERT INTO property_type (id, p_type) OVERRIDING SYSTEM VALUE VALUES (3, 'Room');


--
-- Data for Name: property_usage; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO property_usage (id, p_usage) OVERRIDING SYSTEM VALUE VALUES (1, 'Entire place');
INSERT INTO property_usage (id, p_usage) OVERRIDING SYSTEM VALUE VALUES (2, 'Private room');
INSERT INTO property_usage (id, p_usage) OVERRIDING SYSTEM VALUE VALUES (3, 'Shared room');

-- property combinations

insert into property_combinations( type_id, category_id) values(1, 1);
insert into property_combinations( type_id, category_id) values(1, 2);


insert into property_combinations( type_id, category_id) values(2, 3);
insert into property_combinations( type_id, category_id) values(2, 4);
insert into property_combinations( type_id, category_id) values(2, 5);

insert into property_combinations( type_id, category_id) values(3, 6);
insert into property_combinations( type_id, category_id) values(3, 7);
--
-- Data for Name: properties; Type: TABLE DATA; Schema: public; Owner: postgres
--

insert into subscription_status( s_status, s_description) values('NONE', 'No subscription');