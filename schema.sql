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
CREATE TABLE Events (
    id SERIAL PRIMARY KEY,  -- Auto-incrementing primary key
    name VARCHAR(100) NOT NULL,  -- Event name with a maximum length of 100 characters
    description TEXT,
    total_tickets INT NOT NULL CHECK (total_tickets >= 1),  -- Total tickets, must be at least 1
    -- available_tickets NOT NULL CHECK (available_tickets >= 1)
    booked_tickets INT NOT NULL DEFAULT 0,  -- Booked tickets, defaults to 0
    event_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create the Bookings table
CREATE TABLE Bookings (
    id SERIAL PRIMARY KEY,  -- Auto-incrementing primary key
    user_id INT NOT NULL,  -- Foreign key referencing the Users table
    event_id INT NOT NULL,  -- Foreign key referencing the Events table
    quantity INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES Users(id) ON DELETE CASCADE,  -- Cascade delete when the user is deleted
    FOREIGN KEY (event_id) REFERENCES Events(id) ON DELETE CASCADE  -- Cascade delete when the event is deleted
);
