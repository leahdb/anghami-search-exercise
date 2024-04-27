-- Create 'books' table
CREATE TABLE IF NOT EXISTS books (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    ratings_count INT NOT NULL,
    published_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create 'movies' table
CREATE TABLE IF NOT EXISTS movies (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    director VARCHAR(255) NOT NULL,
    rating FLOAT NOT NULL,
    release_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create 'search_events' table
CREATE TABLE IF NOT EXISTS search_events (
    id INT AUTO_INCREMENT PRIMARY KEY,
    search_id VARCHAR(255) NOT NULL,
    search_query VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL
);

-- Create 'search_clicks' table
CREATE TABLE IF NOT EXISTS search_clicks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    search_id VARCHAR(255) NOT NULL,
    result_type ENUM('book', 'movie') NOT NULL,
    result_id INT NOT NULL,
    result_position INT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
