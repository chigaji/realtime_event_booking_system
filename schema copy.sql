-- schema.sql

-- Create the Users table
CREATE TABLE Users (
    id SERIAL PRIMARY KEY,  -- Auto-incrementing primary key
    username VARCHAR(50) NOT NULL,  -- Username with a maximum length of 50 characters
    password VARCHAR(255) NOT NULL,  -- Password, stored as a hash (usually)
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create the Events table
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    total_tickets INT NOT NULL,
    available_tickets INT NOT NULL,
    event_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create the Bookings table
CREATE TABLE Bookings (
    id SERIAL PRIMARY KEY,  -- Auto-incrementing primary key
    user_id INT NOT NULL,  -- Foreign key referencing the Users table
    event_id INT NOT NULL,  -- Foreign key referencing the Events table
    FOREIGN KEY (user_id) REFERENCES Users(id) ON DELETE CASCADE,  -- Cascade delete when the user is deleted
    FOREIGN KEY (event_id) REFERENCES Events(id) ON DELETE CASCADE  -- Cascade delete when the event is deleted
);
