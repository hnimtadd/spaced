# Spaced

A modern spaced repetition vocabulary learning application built with Go and WebAssembly (WASM). The application uses the FSRS (Free Spaced Repetition Scheduler) algorithm to optimize vocabulary retention through intelligent card scheduling.

## Features

- **Spaced Repetition**: Uses the FSRS algorithm for optimal learning intervals
- **Interactive Flashcards**: Clean, modern card interface with flip animations
- **Audio Pronunciation**: Integrates with Cambridge Dictionary for word pronunciation
- **Session Management**: Track learning sessions and progress statistics
- **WebAssembly Frontend**: Fast, native-like performance in the browser
- **Responsive Design**: Works seamlessly across desktop and mobile devices

## Architecture

### Backend Components
- **HTTP Server**: Serves static files and API endpoints
- **Sound API**: Serverless lambda which fetches pronunciation audio from Cambridge Dictionary
- **Core Models**: Card management, session handling, and FSRS integration
- **HTML Parser**: Parses Cambridge Dictionary pages for audio extraction

### Frontend Components
- **Crafter Framework**: Custom WASM framework for DOM manipulation and state management
- **WebAssembly Modules** (`wasm/`): Go code compiled to WASM for client-side logic
- **Web UI**: HTML, CSS, and JavaScript interface

## Getting Started

### Prerequisites
- Go 1.24 or later
- Make (for build automation)
- Modern web browser with WASM support

### Installation locally

1. Clone the repository:
```bash
git clone https://github.com/hnimtadd/spaced.git
cd spaced
```

2. Build the WASM modules:
```bash
make wasm
```

3. Start the development server:
```bash
make server
```

4. Open your browser and navigate to `http://localhost:8080`

### Development with Nix (Optional)

If you have Nix installed, you can use the provided flake for a consistent development environment:

```bash
# Enter the development shell
make shell

# Or directly with nix
nix develop
```

## Usage

### Starting a Learning Session

1. Visit the home page at `http://localhost:8080`
2. Click "Start Session" to begin learning
3. The system automatically selects cards based on due dates and learning progress

### Learning Interface

- **Card Display**: Shows word, IPA pronunciation, and play button
- **Card Flip**: Click the card to reveal definition and example
- **Rating System**: Rate your recall difficulty:
  - **Again** (Red): Couldn't remember - card will repeat soon
  - **Hard** (Yellow): Difficult to remember - shorter interval
  - **Good** (Green): Normal recall - standard interval
  - **Easy** (Blue): Very easy - longer interval

### Session Management

- Sessions automatically complete when all due cards are reviewed
- Progress is saved in browser localStorage
- View session history and statistics at `/stats`

## Project Structure

```
spaced/
├── api/sound/          # Audio pronunciation API
├── cmd/               # Command line applications
│   ├── server/        # HTTP server
│   └── convert/       # Data conversion utilities
├── src/              # Core Go libraries
│   ├── core/         # Business logic (FSRS, sessions, models)
│   ├── crafter/      # WASM framework
│   ├── html/         # HTML parsing utilities
│   └── utils/        # Common utilities
├── ui/               # Web frontend
│   ├── assets/       # Static assets and WASM binaries
│   ├── *.html        # Page templates
│   ├── style.css     # Styling
│   └── main.js       # JavaScript glue code
├── wasm/             # WASM modules source
│   ├── session/      # Session management WASM
│   └── stats/        # Statistics WASM
├── makefile          # Build automation
├── flake.nix         # Nix development environment
└── vercel.json       # Vercel deployment configuration
```

## Card Data Format

Cards are stored as JSON with the following structure:

```json
{
  "word": "example",
  "ipa": "ɪɡˈzæmpl",
  "definition": "A thing characteristic of its kind or illustrating a general rule.",
  "example": "This is an example sentence."
}
```

## Deployment

### Local Deployment
```bash
make server
```

## Stack

- **Backend**: Go, Gin framework, FSRS algorithm
- **Frontend**: WebAssembly (Go), HTML5, CSS3, JavaScript
- **Build**: Make, Nix (optional)
- **Deployment**: Vercel-ready configuration
- **External APIs**: Cambridge Dictionary (pronunciation)

## License

MIT License
