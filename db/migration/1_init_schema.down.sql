DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS property_reports;
DROP TABLE IF EXISTS user_reports;
DROP TABLE IF EXISTS property_report_reasons;
DROP TABLE IF EXISTS user_report_reasons;
-- trigger
-- function
-- index
DROP TABLE IF EXISTS featured_areas;
DROP TABLE IF EXISTS property_photos;
DROP TABLE IF EXISTS properties;
DROP TABLE IF EXISTS property_combinations;
DROP TABLE IF EXISTS property_category;
DROP TABLE IF EXISTS property_type;
DROP TABLE IF EXISTS property_usage;
DROP TABLE IF EXISTS user_verifications;
DROP TABLE IF EXISTS verification_stage;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS subscription_status;
-- extension
DROP FUNCTION IF EXISTS notify_event;
DROP TRIGGER IF EXISTS users_notify_event;