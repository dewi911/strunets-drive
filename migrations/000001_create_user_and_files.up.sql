CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(255) UNIQUE NOT NULL,
                       password VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE refresh_tokens (
                                id SERIAL PRIMARY KEY,
                                user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                token VARCHAR(255) UNIQUE NOT NULL,
                                expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);


CREATE TABLE files (
                       id VARCHAR(255) PRIMARY KEY,
                       name VARCHAR(255) NOT NULL,
                       path VARCHAR(255) NOT NULL,
                       size BIGINT NOT NULL,
                       username VARCHAR(255) NOT NULL REFERENCES users(username),
                       uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_files_username ON files(username);


CREATE TABLE folders (
                         id VARCHAR(255) PRIMARY KEY,
                         name VARCHAR(255) NOT NULL,
                         parent_id VARCHAR(255),
                         username VARCHAR(255) NOT NULL,
                         created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                         FOREIGN KEY (parent_id) REFERENCES folders(id),
                         FOREIGN KEY (username) REFERENCES users(username)
);

ALTER TABLE files
    ADD COLUMN folder_id VARCHAR(255),
ADD COLUMN is_dir BOOLEAN DEFAULT FALSE,
ADD CONSTRAINT fk_file_folder
    FOREIGN KEY (folder_id)
    REFERENCES folders(id);

INSERT INTO folders (id, name, parent_id, username)
SELECT
    gen_random_uuid()::text,
        'Root',
    NULL,
    username
FROM users;