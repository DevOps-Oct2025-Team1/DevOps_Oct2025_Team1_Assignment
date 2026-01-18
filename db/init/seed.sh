#!/bin/bash

PGPASSWORD=$POSTGRES_PASSWORD psql -U $POSTGRES_USER -d $POSTGRES_DB << EOF
INSERT INTO users (username, password_hash, role) 
VALUES ('$DEFAULT_TESTUSER_USERNAME', '$DEFAULT_TESTUSER_PASSWORD', 'user')
ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash;

INSERT INTO users (username, password_hash, role) 
VALUES ('$DEFAULT_ADMIN_USERNAME', '$DEFAULT_ADMIN_PASSWORD', 'admin')
ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash;
EOF
