DROP TRIGGER IF EXISTS update_masters_modtime ON masters;
DROP INDEX IF EXISTS idx_masters_rating;
DROP INDEX IF EXISTS idx_masters_user_id;
DROP TABLE IF EXISTS masters;