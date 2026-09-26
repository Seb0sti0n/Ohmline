-- The extended explanation shown in the investigation screen (reason is the one-line summary).
ALTER TABLE anomalies ADD COLUMN explanation TEXT NOT NULL DEFAULT '';
