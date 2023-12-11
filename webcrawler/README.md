docker run --name some-postgres -e POSTGRES_PASSWORD=1111 -e POSTGRES_DB=webcrawler -p 5432:5432 -d postgres


CREATE TABLE links (
    id SERIAL PRIMARY KEY,
    link TEXT NOT NULL UNIQUE,
    hash TEXT NOT NULL UNIQUE,
    html TEXT NOT NULL,
    created_at DATE
);
