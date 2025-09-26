# Data Analysis Database Changes

## Overview

The data analysis module has been restructured to use a centralized SQLite database approach instead of managing multiple separate database connections.

## Changes Made

### 1. New Database Architecture

- **Main Database**: `data/data_analysis.db` - A centralized SQLite database that stores all imported flight data
- **Import Process**: When a database is uploaded, all flights and related data are imported into the main database
- **Simplified API**: No more database ID management in the frontend

### 2. New Files

- `data_analysis/database.go`: Contains all database initialization and import functionality
  - `InitMainDatabase()`: Initializes the main SQLite database
  - `ImportFlightsFromDatabase()`: Imports all flights from uploaded databases
  - Import functions for flights, aircraft, position, attitude, and engine data

### 3. Modified Files

- `data_analysis/data_analysis.go`: Simplified to work with the main database only
- `data_analysis/types.go`: Added `SourceID` field to track original flight IDs
- `data_analysis/page.templ`: Updated UI to remove database selection complexity
- `data_analysis/script.templ`: Simplified JavaScript to work with new API
- `main.go`: Added graceful shutdown to properly close database connections

### 4. API Changes

#### Removed Endpoints
- `/data-analysis/databases` - No longer needed
- `/data-analysis/connect` - No longer needed

#### Modified Endpoints
- `/data-analysis/upload` - Now imports flights directly into main database
- `/data-analysis/flights` - Now returns flights from main database only
- `/data-analysis/flight-data` - Simplified, no longer requires `dbId` parameter

#### New Endpoints
- `/data-analysis/api/stats` - Returns main database statistics

## Benefits

1. **Simplified Workflow**: Users upload databases and flights are automatically imported
2. **Persistent Storage**: All flight data is kept in one place, survives server restarts
3. **Better Performance**: No need to manage multiple database connections
4. **Easier Maintenance**: Single database to backup and manage
5. **Data Consolidation**: All flights from different uploads are available in one place

## Usage

1. **Upload a Database**: Drag and drop or select a SQLite flight database file
2. **Automatic Import**: All flights and data are automatically imported into the main database
3. **Select Flight**: Choose from all available flights in the dropdown
4. **Analyze**: Use the visualization tools as before

## Database Schema

The main database uses the same schema as the original flight databases, defined in `data/structure.sql`. This ensures compatibility with existing flight data formats.

## Error Handling

- Database validation ensures uploaded files have the required schema
- Transaction-based imports ensure data integrity
- Proper cleanup of temporary files
- Graceful handling of duplicate or invalid data

## Future Enhancements

- Duplicate flight detection and handling
- Data export functionality
- Advanced search and filtering
- Flight comparison tools
- Database optimization and maintenance tools
