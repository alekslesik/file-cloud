CREATE TABLE
    IF NOT EXISTS tokens (
        hash BINARY (16) PRIMARY KEY,
        user_id INT NOT NULL,
        expiry TIMESTAMP(0) NOT NULL,
        scope TEXT NOT NULL,
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
    );