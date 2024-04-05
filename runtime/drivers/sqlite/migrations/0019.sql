-- Fixes wrong default environment value in migration 0016.sql.
UPDATE instances SET environment = 'prod' WHERE environment = 'production';
