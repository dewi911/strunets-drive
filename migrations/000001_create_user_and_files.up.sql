CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(255) UNIQUE NOT NULL,
                       password VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE folders (
                         id VARCHAR(255) PRIMARY KEY,
                         name VARCHAR(255) NOT NULL,
                         parent_id VARCHAR(255),
                         username VARCHAR(255) NOT NULL,
                         created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                         path_array VARCHAR(255)[] NOT NULL DEFAULT ARRAY[]::VARCHAR[],
                         FOREIGN KEY (parent_id) REFERENCES folders(id) ON DELETE CASCADE,
                         FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE
);

CREATE INDEX idx_folders_parent_id ON folders(parent_id);
CREATE INDEX idx_folders_username ON folders(username);
CREATE INDEX idx_folders_path_array ON folders USING gin(path_array);

CREATE TABLE files (
                       id VARCHAR(255) PRIMARY KEY,
                       name VARCHAR(255) NOT NULL,
                       path VARCHAR(255) NOT NULL,
                       size BIGINT NOT NULL,
                       username VARCHAR(255) NOT NULL,
                       folder_id VARCHAR(255) NOT NULL,
                       is_dir BOOLEAN DEFAULT FALSE,
                       uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                       FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE,
                       FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

CREATE INDEX idx_files_username ON files(username);
CREATE INDEX idx_files_folder_id ON files(folder_id);


CREATE TABLE refresh_tokens (
                                id SERIAL PRIMARY KEY,
                                user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                token VARCHAR(255) UNIQUE NOT NULL,
                                expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);

CREATE OR REPLACE FUNCTION create_root_folder()
RETURNS TRIGGER AS $$
BEGIN
INSERT INTO folders (id, name, parent_id, username, path_array)
VALUES (gen_random_uuid()::text, 'Root', NULL, NEW.username, ARRAY[]::VARCHAR[]);
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER create_root_folder_trigger
    AFTER INSERT ON users
    FOR EACH ROW
    EXECUTE FUNCTION create_root_folder();