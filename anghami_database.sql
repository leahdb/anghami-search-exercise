-- Create 'books' table
CREATE TABLE IF NOT EXISTS books (
    bookID VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    authors VARCHAR(255) NOT NULL,
    average_rating VARCHAR(255) NULL,
    isbn VARCHAR(255) NULL,
    isbn13 VARCHAR(255) NULL,
    language_code VARCHAR(255) NULL,
    num_pages VARCHAR(255) NULL,
    ratings_count INT NOT NULL,
    text_reviews_count VARCHAR(255) NULL,
    publication_date VARCHAR(255) NULL,
    publisher VARCHAR(255) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Create 'movies' table
CREATE TABLE IF NOT EXISTS movies (
    movieID INT AUTO_INCREMENT PRIMARY KEY,
    Title VARCHAR(255) NULL,
    Year VARCHAR(255) NULL,
    Summary VARCHAR(255) NULL,
    Short_Summary VARCHAR(255) NULL,
    IMDB_ID VARCHAR(255) NULL,
    Runtime VARCHAR(255) NULL,
    YouTube_Trailer VARCHAR(255) NULL,
    Rating VARCHAR(255) NULL,
    Movie_Poster VARCHAR(255) NULL,
    Director VARCHAR(255) NULL,
    Writers VARCHAR(255) NULL,
    Cast VARCHAR(255) NULL,
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
