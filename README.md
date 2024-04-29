# Anghami Search Exercise

## Description
This project demonstrates my implementation of a search functionality using Go (Golang) and MySQL. It includes functionalities for importing data from CSV files into the database, handling search queries, tracking search events and click events, generating insights from click data, and more.
## Prerequisites
Before running the project, ensure that you have Docker and Docker Compose installed on your system.

## Installation
To run the project, follow these steps:

1. **Clone the Repository**:
   
```
git clone https://github.com/leahdb/anghami-search-exercise.git
```


2. **Navigate to the Project Directory**:

```
cd anghami-search-exercise
```


3. **Copy the Environment File**:
- Copy the .env.example file and rename it to .env:

  ```
  cp .env.example .env
  ```

4. **Fill in Environment Variables**:
- Open the .env file in a text editor and fill in the required environment variables with your desired values.


5. **Start Docker Containers**:
  
  ```
  docker-compose up -d
  ```


6. **Verify Containers are Running**:
  
  ```
  docker ps
  ```

Ensure that both MySQL container is running.

7. **Access the Application**:
- MySQL: Access the MySQL database using your preferred MySQL client or command-line tool.

8. **Run the Application**:
- Start the Go server by running the following command:
  ```
  go run main.go
  ```

  
## Postman Collection
You can use the provided Postman collection to explore and interact with all the endpoints of the Anghami Search Exercise API.

[Download Postman Collection](https://www.postman.com/leahdb/workspace/anghami-endpoints-collection/collection/15530394-c1ca7f2f-4c7a-4144-85dd-80fa3425e1ed?action=share&creator=15530394)



## Usage
Once the application is running, you can interact with it using the following endpoints:
- `/search`: Handle search queries.
- `/report-search`: Report search events.
- `/report-click`: Report click events.
- `/import-books`: Import data from the books.csv file into the 'books' table.
- `/import-movies`: Import data from the movies.csv file into the 'movies' table.
- `/generate-insights`: Generate insights from click data.
