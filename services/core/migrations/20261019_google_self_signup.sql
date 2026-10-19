BEGIN;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS google_sub VARCHAR(255),
    ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(24) NOT NULL DEFAULT 'password',
    ADD COLUMN IF NOT EXISTS terms_accepted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS privacy_acknowledged_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS terms_version VARCHAR(32),
    ADD COLUMN IF NOT EXISTS privacy_notice_version VARCHAR(32);

CREATE UNIQUE INDEX IF NOT EXISTS users_google_sub_uidx
    ON users (google_sub)
    WHERE google_sub IS NOT NULL;

-- This index intentionally fails if legacy data contains case-variant duplicate
-- emails; resolve those records before applying the migration, never silently merge.
CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_uidx
    ON users (LOWER(email));

-- The form accepts punctuation for readability; canonical digits remain unique.
-- Resolve pre-existing normalized collisions before applying this migration.
CREATE UNIQUE INDEX IF NOT EXISTS companies_nit_canonical_uidx
    ON companies (REGEXP_REPLACE(nit, '[^0-9]', '', 'g'));

COMMIT;
