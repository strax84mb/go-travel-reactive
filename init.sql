DROP TABLE IF EXISTS routes;
DROP TABLE IF EXISTS airports;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS cities;
DROP INDEX idx_users_username on users;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
                       id INTEGER NOT NULL AUTO_INCREMENT,
                       username VARCHAR(30) UNIQUE NOT NULL,
                       `password` VARCHAR(200) NOT NULL,
                       salt VARCHAR(100) NOT NULL,
                       `role` VARCHAR(15) NOT NULL,
                       PRIMARY KEY (id)
);

INSERT INTO users (username, `password`, salt, `role`) VALUES
('admin',
 '92766f4ade6d45666bd4c26798a39c974874c118bd4d95815b62a548988fd7db33060246e2555bf93328d5dfabd3ffbd799efafbe4b9e775ca46d005fe0932072857d6e63173fa2c41ccd10d194bfef8',
 '92766f4ade6d45666bd4c26798a39c97',
 'ADMIN');

CREATE TABLE cities (
                        id INTEGER NOT NULL AUTO_INCREMENT,
                        `name` VARCHAR(100) NOT NULL,
                        country VARCHAR(100) NOT NULL,
                        PRIMARY KEY (id)
);


CREATE TABLE comments (
                          id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
                          city_id INTEGER NOT NULL,
                          poster_id INTEGER NOT NULL,
                          `text` VARCHAR(255) NOT NULL,
                          created DATETIME NOT NULL,
                          modified DATETIME NOT NULL,
                          FOREIGN KEY (city_id) REFERENCES cities(id) ON DELETE CASCADE,
                          FOREIGN KEY (poster_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE airports (
                          id INTEGER NOT NULL AUTO_INCREMENT,
                          airport_id INTEGER UNIQUE NOT NULL,
                          `name` VARCHAR(100) NOT NULL,
                          city_id INTEGER NOT NULL,
                          FOREIGN KEY (city_id) REFERENCES cities(id),
                          PRIMARY KEY (id)
);

CREATE TABLE routes (
                        id INTEGER NOT NULL AUTO_INCREMENT,
                        source_id INTEGER NOT NULL,
                        destination_id INTEGER NOT NULL,
                        price REAL NOT NULL,
                        FOREIGN KEY (source_id) REFERENCES airports(id),
                        FOREIGN KEY (destination_id) REFERENCES airports(id),
                        PRIMARY KEY (id)
);