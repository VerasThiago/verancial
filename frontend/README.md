# Verancial Frontend

A React TypeScript frontend for the Verancial financial data management platform.

## Features

- **Secure Authentication**: JWT-based authentication with automatic token management
- **Dashboard**: View connected bank accounts and transaction statistics
- **Modern UI**: Clean, responsive design with smooth animations
- **Real-time Data**: Shows how outdated your financial data is
- **Bank Account Management**: Connect and manage multiple bank accounts

## Getting Started

### Prerequisites

- Node.js 16+ 
- npm or yarn

### Installation

1. Install dependencies:
```bash
npm install
```

2. Create environment file:
```bash
cp .env.example .env
```

3. Update environment variables in `.env`:
```
REACT_APP_API_URL=http://localhost:8080/api/v0
REACT_APP_LOGIN_URL=http://localhost:8081
PORT=3000
```

### Development

Start the development server:
```bash
npm start
```

The app will be available at `http://localhost:3000`

### Building for Production

```bash
npm run build
```

### Docker

Build and run with Docker:
```bash
docker build -t verancial-frontend .
docker run -p 3000:3000 verancial-frontend
```

Or use docker-compose from the project root:
```bash
make start_frontend_docker
```

## API Integration

The frontend communicates with:
- **Login Service** (port 8081): User authentication
- **API Service** (port 8080): Dashboard data and bank account management

## Security Features

- JWT tokens stored in localStorage with automatic cleanup
- Automatic redirect to login on authentication failure
- CSRF protection headers
- Content Security Policy headers

## Architecture

```
src/
├── components/          # React components
│   ├── Login.tsx       # Login form
│   └── Dashboard.tsx   # Main dashboard
├── services/           # API services
│   └── api.ts         # HTTP client with auth
├── styles/            # CSS files
│   ├── App.css
│   ├── Login.css
│   └── Dashboard.css
└── App.tsx            # Main app with routing
```

## Environment Variables

- `REACT_APP_API_URL`: Backend API base URL
- `REACT_APP_LOGIN_URL`: Login service base URL  
- `PORT`: Development server port 