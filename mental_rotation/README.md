# Mental Rotation Package

The `mental_rotation` package implements a computerized mental rotation test for psychological research, specifically designed to assess spatial cognitive abilities. It presents 3D object rotation tasks and records participant responses with precise timing data.

## Overview

This package provides:
- A complete mental rotation testing interface
- Automated task presentation from embedded image sets
- Precise response timing measurement
- Result storage and analysis capabilities
- German language support for instructions

## Key Features

### Mental Rotation Testing
- **3D Object Tasks**: Presents pairs of 3D objects for comparison
- **Rotation Detection**: Participants determine if objects are the same or different
- **Automatic Scoring**: Validates responses against correct answers
- **Timing Analysis**: Records precise response times for each task

### Task Management
- **Image Discovery**: Automatically detects all JPG images in embedded directory
- **Task Generation**: Creates tasks based on filename conventions
- **Sequential Presentation**: Presents tasks in consistent order
- **Progress Tracking**: Manages progression through task sequence

### Data Collection
- **Participant Identification**: Requires participant ID for data tracking
- **Response Recording**: Captures correctness and reaction times
- **Persistent Storage**: Saves results to JSON file for analysis
- **Session Management**: Handles complete testing sessions

## Architecture

### Core Components

**`mental_rotation.go`**
- Main package logic and HTTP handlers
- Task generation from embedded images
- Result storage and retrieval
- Thread-safe operations with mutex protection

**`mental_rotation.html`**
- Complete web-based testing interface
- German language instructions and UI
- Responsive design for various screen sizes
- Interactive task presentation with timing

## Data Structures

### Task
```go
type Task struct {
    ID            int       `json:"id"`
    Image         string    `json:"image"`
    CorrectAnswer bool      `json:"correctAnswer"`
    StartTime     time.Time `json:"startTime"`
    EndTime       time.Time `json:"endTime"`
}
```

### Result
```go
type Result struct {
    ParticipantID string        `json:"participantId"`
    Image         string        `json:"image"`
    IsCorrect     bool          `json:"isCorrect"`
    TimeTaken     time.Duration `json:"timeTaken"`
    Timestamp     string        `json:"timestamp"`
}
```

## Image Naming Convention

The package uses filename conventions to determine correct answers:
- **Same Object**: Standard naming (e.g., `1_0.jpg`, `2_50.jpg`)
- **Different Object**: Filenames ending with `_R.jpg` (e.g., `1_50_R.jpg`, `24_150_R.jpg`)

The `_R` suffix indicates a reflected/mirrored version, making it a different object that cannot be achieved through rotation alone.

## API Endpoints

### GET `/mental-rotation/tasks`
Retrieve all available tasks.

**Response:**
```json
[
  {
    "id": 1,
    "image": "1_0.jpg",
    "correctAnswer": true
  },
  {
    "id": 2,
    "image": "1_50_R.jpg",
    "correctAnswer": false
  }
]
```

### POST `/mental-rotation/submit`
Submit a task result.

**Request Body:**
```json
{
  "participantId": "P001",
  "image": "1_0.jpg",
  "isCorrect": true,
  "timeTaken": 2500,
  "timestamp": "2025-06-03T10:30:45.123Z"
}
```

### GET `/mental-rotation/results`
Retrieve all recorded results.

**Response:**
```json
[
  {
    "participantId": "P001",
    "image": "1_0.jpg",
    "isCorrect": true,
    "timeTaken": 2500000000,
    "timestamp": "2025-06-03T10:30:45.123Z"
  }
]
```

### GET `/mental-rotation/images/[filename]`
Serve embedded image files for task presentation.

### GET `/mental-rotation`
Serve the main testing interface.

## Test Protocol

### Instructions (German)
The interface provides German instructions:
> "In dieser Aufgabe werden zwei 3D-Grafiken angezeigt. Ihre Aufgabe ist es, festzustellen, ob die beiden Grafiken dieselbe 3D-Figur darstellen, obwohl sie in unterschiedlichen Orientierungen gezeichnet sind."

### Response Options
- **"Gleiche Figur"** (Same Figure): Objects represent the same 3D shape
- **"Andere Figur"** (Different Figure): Objects represent different shapes

### Task Flow
1. Participant enters ID
2. Instructions are displayed
3. Sequential presentation of tasks
4. Response recording with timing
5. Completion message

## Data Storage

Results are stored in:
```
data/mental_rotation_results.json
```

**Storage Format:**
```json
[
  {
    "participantId": "P001",
    "image": "1_0.jpg",
    "isCorrect": true,
    "timeTaken": 2500000000,
    "timestamp": "2025-06-03T10:30:45.123Z"
  }
]
```

## Image Set

The package includes embedded 3D object images with various rotations:
- Multiple object types (numbered series)
- Different rotation angles (0°, 50°, 100°, 150°)
- Both normal and reflected versions
- Consistent visual presentation

## Usage Examples

### Starting a Test Session
```javascript
// Set participant ID and start test
function startTask() {
    const participantId = document.getElementById('participant-id').value;
    if (!participantId) {
        // Show error message
        return;
    }
    
    // Load tasks and begin testing
    loadTasks().then(() => {
        showCurrentTask();
    });
}
```

### Recording Responses
```javascript
// Submit participant response
async function submitAnswer(userAnswer) {
    const result = {
        participantId: participantId,
        image: tasks[currentTaskIndex].image,
        isCorrect: userAnswer === tasks[currentTaskIndex].correctAnswer,
        timeTaken: endTime - startTime,
        timestamp: new Date().toISOString()
    };
    
    await fetch('/mental-rotation/submit', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(result)
    });
}
```

## Performance Metrics

The system captures:
- **Accuracy**: Percentage of correct responses
- **Response Time**: Millisecond precision timing
- **Error Patterns**: Analysis of incorrect responses
- **Individual Differences**: Participant-specific performance

## Research Applications

### Spatial Ability Assessment
- Measures mental rotation capabilities
- Identifies individual differences in spatial cognition
- Provides baseline cognitive measures for research

### Experimental Design
- Randomizable task presentation
- Controlled stimulus presentation
- Standardized testing protocol
- Reliable timing measurements

### Data Analysis Support
- Structured data output for statistical analysis
- Session-based data organization
- Participant tracking capabilities
- Result export functionality

## Interface Features

### Responsive Design
- Adapts to different screen sizes
- Optimized for desktop and tablet use
- Clear visual presentation of 3D objects
- Intuitive button layout

### User Experience
- Clear German instructions
- Progress indication
- Error prevention and validation
- Completion confirmation

### Accessibility
- High contrast visual elements
- Large, clear buttons
- Keyboard navigation support
- Clear typography

## Configuration

### Embedded Resources
- Images embedded in binary for portability
- No external dependencies for image serving
- Automatic image discovery and task generation

### File System
- Automatic creation of data directory
- JSON-based result storage
- Error handling for file operations

## Error Handling

Comprehensive error handling for:
- Invalid participant IDs
- Network request failures
- File system access issues
- JSON parsing errors
- Missing or corrupted images

## Thread Safety

- Mutex protection for result storage
- Concurrent access handling for multiple participants
- Atomic operations for data consistency

## Integration

The package integrates with:
- Main application routing system
- Shared data directory structure
- Common error handling patterns
- Consistent API design

This package provides a complete solution for mental rotation testing in psychological research, with robust data collection and analysis capabilities.
