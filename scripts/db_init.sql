drop table IF EXISTS ipv4_geolocation_map;
drop table IF EXISTS ipv6_geolocation_map;
drop table IF EXISTS geolocation_country_map;

CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Table for IPv4 CIDR ranges
CREATE TABLE ipv4_geolocation_map (
id SERIAL PRIMARY KEY,
cidr_ip CIDR NOT NULL,         -- stores IPv4 CIDR (e.g., 163.116.243.64/27)
geoname_id INTEGER NOT NULL,
created_at TIMESTAMPTZ DEFAULT now()
);

-- I had no idea a CIDR type existed and was prepared to convert ips->decimal, create min/max columns, and query w/ "BETWEEN"
-- However, AI suggested a CIDR type when I asked the best data type for the scenario of searching for a single ip in a range notated by ip/mask
-- I wasn't sure how to efficiently handle converting ipv6 to decimal in a db since int size couldn't hold that big a value
-- This led to further discovery of a Generalized Search Tree, which is useful in speeding up network containment queries for CIDR types
-- It will speed up scanning within the range when using the << operator
CREATE INDEX idx_ip4_cidr_ip ON ipv4_geolocation_map USING gist (cidr_ip);

-- Table for IPv6 CIDR ranges
CREATE TABLE ipv6_geolocation_map (
id SERIAL PRIMARY KEY,
cidr_ip CIDR NOT NULL,         -- stores IPv6 CIDR (e.g., 2001:218:4000:20::/59)
geoname_id INTEGER NOT NULL,
created_at TIMESTAMPTZ DEFAULT now()
);

-- I had no idea a CIDR type existed and was prepared to convert ips->decimal, create min/max columns, and query w/ "BETWEEN"
-- However, AI suggested a CIDR type when I asked the best data type for the scenario of searching for a single ip in a range notated by ip/mask
-- I wasn't sure how to efficiently handle converting ipv6 to decimal in a db since int size couldn't hold that big a value
-- This led to further discovery of a Generalized Search Tree, which is useful in speeding up network containment queries for CIDR types
-- It will speed up scanning within the range when using the << operator
CREATE INDEX idx_ip6_cidr_ip ON ipv6_geolocation_map USING gist (cidr_ip);

-- Table mapping geolocation codes to country names
CREATE TABLE geolocation_country_map (
geoname_id INTEGER PRIMARY KEY,
country_name TEXT NOT NULL
);

-- Sample data for IPv4
INSERT INTO ipv4_geolocation_map (cidr_ip, geoname_id) VALUES
('163.116.243.64/27', 2921044),
('163.116.242.120/30', 3017382);

-- Sample data for IPv6
INSERT INTO ipv6_geolocation_map (cidr_ip, geoname_id) VALUES
('2001:218:4000:20::/59', 226074),
('2001:200::/32', 1861060);

-- Sample geolocation mappings
INSERT INTO geolocation_country_map (geoname_id, country_name) VALUES
(3017382, 'France'),
(2921044, 'Germany'),
(226074, 'Uganda'),
(1861060, 'Japan');

