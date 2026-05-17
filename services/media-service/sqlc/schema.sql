CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE records (
  id UUID PRIMARY KEY,
  filename TEXT NOT NULL,
  mime_type TEXT NOT NULL,
  size BIGINT NOT NULL,
  path TEXT NOT NULL,
  user_id UUID NOT NULL,
  uploaded_at TIMESTAMP NOT NULL
);