-- Drop in reverse FK dependency order

DROP TRIGGER IF EXISTS update_clients_modtime ON clients;
DROP TABLE IF EXISTS clients;

DROP INDEX IF EXISTS idx_portfolio_photos_master_id;
DROP TABLE IF EXISTS master_portfolio_photos;

DROP TRIGGER IF EXISTS update_appointments_modtime ON appointments;
DROP TABLE IF EXISTS appointments;

DROP TRIGGER IF EXISTS update_work_intervals_modtime ON master_work_intervals;
DROP INDEX IF EXISTS idx_work_intervals_master_date;
DROP TABLE IF EXISTS master_work_intervals;

DROP TRIGGER IF EXISTS update_master_settings_modtime ON master_settings;
DROP TABLE IF EXISTS master_settings;

DROP TRIGGER IF EXISTS update_master_services_modtime ON master_services;
DROP TABLE IF EXISTS master_services;

DROP TRIGGER IF EXISTS update_masters_modtime ON masters;
DROP INDEX IF EXISTS idx_masters_category_rating;
DROP INDEX IF EXISTS idx_masters_rating;
DROP TABLE IF EXISTS masters;

DROP TRIGGER IF EXISTS update_category_modtime ON category;
DROP TABLE IF EXISTS category;

DROP INDEX IF EXISTS idx_user_email;
DROP TABLE IF EXISTS "user";

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP EXTENSION IF EXISTS btree_gist;
DROP EXTENSION IF EXISTS "uuid-ossp";
