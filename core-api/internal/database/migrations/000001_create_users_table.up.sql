
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


CREATE TABLE users (
                        id text PRIMARY KEY NOT NULL,
                        first_name VARCHAR(100) NOT NULL,
                        last_name VARCHAR(100) NOT NULL,
                        username VARCHAR(50) UNIQUE NOT NULL,
                        google_email VARCHAR(255),
                        apple_email VARCHAR(255),
                        refresh_token VARCHAR(255),
                        refresh_token_expiry_date TIMESTAMPTZ,
                        avatar_url VARCHAR(512),
                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
                        monthly_saving_goal NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
                        preferred_currency CHAR(3) DEFAULT 'MKD',

                        CONSTRAINT at_least_an_email_to_be_present CHECK (
                            google_email IS NOT NULL OR
                            apple_email IS NOT NULL
                        )
);

CREATE UNIQUE INDEX idx_unique_user_email
    ON users(COALESCE(google_email, apple_email))
    WHERE google_email IS NOT NULL OR apple_email IS NOT NULL;

CREATE INDEX idx_users_emails_partial
    ON users(COALESCE(google_email, apple_email))
    WHERE google_email IS NOT NULL OR apple_email IS NOT NULL;

CREATE INDEX idx_users_refresh_token
    ON users(refresh_token)
    WHERE refresh_token IS NOT NULL;

