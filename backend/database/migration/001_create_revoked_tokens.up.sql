-- K-1: blocklist jti JWT praktikum. Token yang sudah logout dicabut sampai exp-nya,
-- lalu baris boleh dibuang (tak permanen). Additive: hanya CREATE TABLE, tidak mengubah tabel ada.
CREATE TABLE IF NOT EXISTS revoked_tokens (
    jti        varchar(64) PRIMARY KEY,
    user_id    integer     NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_revoked_tokens_user_id ON revoked_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_revoked_tokens_expires_at ON revoked_tokens(expires_at);
