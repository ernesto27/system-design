-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE hotel (
    hotel_id INT PRIMARY KEY,
    name VARCHAR(255),
    address VARCHAR(255),
    location VARCHAR(255)
);

CREATE TABLE room (
    room_id INT PRIMARY KEY,
    floor INT,
    number INT,
    price DECIMAL(10, 2),
    hotel_id INT,
    name VARCHAR(255),
    FOREIGN KEY (hotel_id) REFERENCES hotel(hotel_id)
);

CREATE TABLE room_rate (
    hotel_id INT,
    date DATE,
    rate DECIMAL(10, 2),
    PRIMARY KEY (hotel_id, date),
    FOREIGN KEY (hotel_id) REFERENCES hotel(hotel_id)
);

CREATE TABLE guest (
    guest_id INT PRIMARY KEY,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    email VARCHAR(255),
    reservation_token VARCHAR(100)
);


CREATE TABLE room_table_inventory (
    hotel_id INT,
    room_id INT,
    date DATE,
    total_inventory INT,
    total_reserved INT,
    PRIMARY KEY (hotel_id, room_id, date),
    FOREIGN KEY (hotel_id) REFERENCES hotel(hotel_id),
    FOREIGN KEY (room_id) REFERENCES room(room_id)
);

CREATE TABLE reservation (
    reservation_id INT PRIMARY KEY,
    hotel_id INT,
    room_id INT,
    start_date DATE,
    end_date DATE,
    status VARCHAR(50),
    guest_id INT,
    version INT,
    FOREIGN KEY (hotel_id) REFERENCES hotel(hotel_id),
    FOREIGN KEY (room_id) REFERENCES room(room_id),
    FOREIGN KEY (guest_id) REFERENCES guest(guest_id)
);

-- Inserts for hotel table
INSERT INTO hotel (hotel_id, name, address, location)
VALUES (1, 'Hotel A', '123 Main St', 'New York'),
       (2, 'Hotel B', '456 Elm St', 'Los Angeles'),
       (3, 'Hotel C', '789 Oak St', 'Chicago');

-- Inserts for room table
INSERT INTO room (room_id, floor, number, price, hotel_id, name)
VALUES (1, 1, 101, 100.00, 1, 'Standard Room'),
    (2, 1, 102, 100.00, 1, 'Standard Room'),
    (3, 2, 201, 150.00, 2, 'Deluxe Room'),
    (4, 2, 202, 150.00, 2, 'Deluxe Room'),
    (5, 3, 301, 200.00, 3, 'Suite');

-- Inserts for room_rate table
INSERT INTO room_rate (hotel_id, date, rate)
VALUES (1, '2022-01-01', 100.00),
       (1, '2022-01-02', 120.00),
       (2, '2022-01-01', 150.00),
       (2, '2022-01-02', 180.00),
       (3, '2022-01-01', 200.00),
       (3, '2022-01-02', 220.00);

-- Inserts for guest table
INSERT INTO guest (guest_id, first_name, last_name, email, reservation_token)
VALUES (1, 'John', 'Doe', 'john.doe@example.com', '1234'),
       (2, 'Jane', 'Smith', 'jane.smith@example.com', ''),
       (3, 'Mike', 'Johnson', 'mike.johnson@example.com', '');

-- Inserts for room_table_inventory table
INSERT INTO room_table_inventory (hotel_id, room_id, date, total_inventory, total_reserved)
VALUES (1, 1, '2022-01-01', 10, 5),
       (1, 1, '2022-01-02', 10, 3),
       (2, 3, '2022-01-01', 5, 2),
       (2, 3, '2022-01-02', 5, 1),
       (3, 5, '2022-01-01', 3, 1),
       (3, 5, '2022-01-02', 3, 0);

-- Inserts for reservation table
INSERT INTO reservation (reservation_id, hotel_id, room_id, start_date, end_date, status, guest_id, version)
VALUES (1, 1, 1, '2022-01-01', '2022-01-02', 'Confirmed', 1, 1),
       (2, 1, 2, '2022-01-01', '2022-01-02', 'Confirmed', 2, 1),
       (3, 2, 3, '2022-01-01', '2022-01-02', 'Confirmed', 3, 1);


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE reservation;
DROP TABLE room_table_inventory;
DROP TABLE guest;
DROP TABLE room_rate;
DROP TABLE room;
DROP TABLE hotel;
