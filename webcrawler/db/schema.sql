CREATE TABLE links (
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    hash TEXT NOT NULL UNIQUE,
    html TEXT NOT NULL,
    keywords TEXT, 
    description TEXT,
    created_at DATE
);


CREATE TABLE IF NOT EXISTS images (
	id SERIAL PRIMARY KEY,
	url TEXT NOT NULL,
    url_image TEXT NOT NULL UNIQUE,
	path TEXT,
    hash TEXT NOT NULL UNIQUE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)