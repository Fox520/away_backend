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
