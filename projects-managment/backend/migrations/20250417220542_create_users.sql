-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

INSERT INTO users (username, email, password_hash) VALUES
('admin', 'test@gmail.com', '$2a$10$l0zmxD/27jlh9zDJ7aXbCeZqQD4IhgXsIMzdWtDyb22rq9Bh7lElO'),
('seller_user', 'seller@example.com', '$2a$10$l0zmxD/27jlh9zDJ7aXbCeZqQD4IhgXsIMzdWtDyb22rq9Bh7lElO'),
('manager_user', 'manager@example.com', '$2a$10$l0zmxD/27jlh9zDJ7aXbCeZqQD4IhgXsIMzdWtDyb22rq9Bh7lElO'),
('dev_user', 'developer@example.com', '$2a$10$l0zmxD/27jlh9zDJ7aXbCeZqQD4IhgXsIMzdWtDyb22rq9Bh7lElO'),
('ux_user', 'ux@example.com', '$2a$10$l0zmxD/27jlh9zDJ7aXbCeZqQD4IhgXsIMzdWtDyb22rq9Bh7lElO');

-- Create roles table
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

-- Insert predefined roles
INSERT INTO roles (name) VALUES
('seller'),
('manager'),
('developer'),
('ux');

-- Create user_roles junction table
CREATE TABLE user_roles (
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- Create project_statuses table
CREATE TABLE project_statuses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

-- Insert default project statuses
INSERT INTO project_statuses (name) VALUES
('pending'),
('in_progress'),
('completed'),
('cancelled');

-- Create projects table
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status_id INT NOT NULL,
    created_user_id INT NOT NULL,
    time_estimation INT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (status_id) REFERENCES project_statuses(id),
    FOREIGN KEY (created_user_id) REFERENCES users(id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS project_statuses;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
