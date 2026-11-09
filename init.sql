-- File for initialising our database.
CREATE TABLE stored_files (
        Id VARCHAR(255),
        file_name VARCHAR(255) NOT NULL,
        file_path VARCHAR(512) NOT NULL UNIQUE,
        PRIMARY KEY (Id)
);