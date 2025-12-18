-- MySQL Initialization Script
-- This runs automatically when the MySQL container starts for the first time

-- Create the ip2country table
CREATE TABLE IF NOT EXISTS ip2country (
    ip VARCHAR(45) PRIMARY KEY,          -- Supports both IPv4 and IPv6
    city VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    INDEX idx_ip (ip)                    -- Index for fast lookups
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Insert sample data (we'll add more later)
INSERT INTO ip2country (ip, city, country) VALUES
    ('8.8.8.8', 'Mountain View', 'United States'),
    ('1.1.1.1', 'Sydney', 'Australia'),
    ('2.22.233.255', 'London', 'United Kingdom')
ON DUPLICATE KEY UPDATE city=VALUES(city), country=VALUES(country);

-- Log successful initialization
SELECT 'MySQL database initialized successfully!' AS message;
