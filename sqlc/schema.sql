CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE records (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  filename TEXT NOT NULL,
  mime_type TEXT NOT NULL,
  size BIGINT NOT NULL,
  path TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  uploaded_at TIMESTAMP NOT NULL DEFAULT NOW()
);