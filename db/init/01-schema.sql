-- Create users table for authentication
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample users
INSERT INTO users (username, password_hash, role) VALUES
('umamusume', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user') ON CONFLICT (username) DO NOTHING;

INSERT INTO users (username, password_hash, role) VALUES
('admin', '$2a$10$SY3yH2XpavczmRrQXbXYkuXD8mVVCN7M8fetmb.tk32XbO18TnqQ.', 'admin') ON CONFLICT (username) DO NOTHING;

-- Note: Replace the hash with a real bcrypt hash for 'password'
