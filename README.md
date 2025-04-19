# Enterprise Health Appointment Scheduling System (EHASS)

EHASS is a comprehensive healthcare appointment scheduling platform built with Go, designed to streamline the management of medical appointments between patients and healthcare providers.

## Overview

EHASS provides a robust API for healthcare facilities to manage patient appointments, doctor schedules, and medical records. The system supports authentication, role-based access control, and real-time availability management.

## Features

- **User Management**: Registration, authentication, and profile management for patients and healthcare providers
- **Doctor Management**: Specialty classification, availability scheduling, and profile management
- **Patient Management**: Medical history, appointment tracking, and profile settings
- **Appointment Scheduling**: Real-time availability checking, appointment creation, modification, and cancellation
- **Medical Records**: Secure storage and retrieval of patient medical information
- **Audit Logging**: Comprehensive activity tracking for compliance and security
- **Enhanced Authentication**: OAuth integration with GitHub/Google, email verification, 2FA, and password reset

## Installation

### Prerequisites

- Go 1.20 or higher
- PostgreSQL 14.0 or higher
- Docker and Docker Compose (for containerized deployment)

### Local Development Setup

1. Clone the repository:
```bash
git clone https://github.com/whitewalker-sa/ehass.git
cd ehass
```

2. Install dependencies:
```bash
go mod download
```

3. Configure the environment:
```bash
cp configs/config.example.yaml configs/config.yaml
# Edit config.yaml with your database credentials and other settings
```

4. Run the database and apply migrations:
```bash
make run-db
```

5. Run the application:
```bash
make run-dev
```

### Docker Setup

1. Build and start the containers:
```bash
docker-compose up -d
```

The application will be available at `http://localhost:8080`

## Database Migrations

EHASS includes a built-in migration system to manage database schema changes:

### Running Migrations

Migrations are automatically applied when starting the database with `make run-db`. You can also run them manually:

```bash
make migrate
```

### Creating New Migrations

To create a new migration file:

```bash
make migration-create
```

This will prompt for a migration name and generate a timestamped Go file in the `internal/migrations` directory.

### Rolling Back Migrations

To roll back the last applied migration:

```bash
make migrate-rollback
```

## API Documentation

Once the server is running, API documentation is available at:
- Swagger UI: `http://localhost:8080/swagger/index.html`
- OpenAPI JSON: `http://localhost:8080/swagger/doc.json`

### Key Endpoints

#### Authentication
- `POST /api/v1/auth/register`: Register a new user
- `POST /api/v1/auth/login`: Authenticate and get access token
- `POST /api/v1/auth/oauth/login`: Login with OAuth provider (GitHub/Google)
- `POST /api/v1/auth/verify-email`: Verify user email address
- `POST /api/v1/auth/request-password-reset`: Request password reset
- `POST /api/v1/auth/reset-password`: Reset password with token
- `POST /api/v1/auth/refresh-token`: Get new access token using refresh token
- `POST /api/v1/auth/verify-2fa`: Verify two-factor authentication code
- `POST /api/v1/auth/logout`: Invalidate current session

#### Authentication Management (Protected Routes)
- `POST /api/v1/auth/setup-2fa`: Set up two-factor authentication
- `POST /api/v1/auth/enable-2fa`: Enable two-factor authentication
- `POST /api/v1/auth/disable-2fa`: Disable two-factor authentication
- `POST /api/v1/auth/link-oauth`: Link OAuth provider to account

#### User Management
- `GET /api/v1/users/{id}`: Get user details
- `PUT /api/v1/users/{id}`: Update user information
- `DELETE /api/v1/users/{id}`: Delete user account
- `PUT /api/v1/users/{id}/change-password`: Change user password
- `PUT /api/v1/users/{id}/avatar`: Update user avatar

#### Doctor Management
- `POST /api/v1/doctors`: Create doctor profile
- `GET /api/v1/doctors`: List all doctors
- `GET /api/v1/doctors/{id}`: Get doctor details
- `PUT /api/v1/doctors/{id}`: Update doctor information
- `GET /api/v1/doctors/specialty/{specialty}`: Find doctors by specialty
- `GET /api/v1/doctors/user/{userID}`: Get doctor by user ID

#### Patient Management
- `POST /api/v1/patients`: Create patient profile
- `GET /api/v1/patients/{id}`: Get patient details
- `PUT /api/v1/patients/{id}`: Update patient information
- `GET /api/v1/patients/user/{userID}`: Get patient by user ID

#### Appointment Management
- `POST /api/v1/appointments`: Create a new appointment
- `GET /api/v1/appointments/{id}`: Get appointment details
- `GET /api/v1/appointments/doctor/{doctorId}`: List doctor's appointments
- `GET /api/v1/appointments/patient/{patientId}`: List patient's appointments
- `PUT /api/v1/appointments/{id}`: Update appointment
- `DELETE /api/v1/appointments/{id}`: Cancel appointment

## Project Structure

```
ehass/
├── cmd/
│   └── server/             # Application entry point
├── configs/                # Configuration files
├── deployments/            # Deployment configurations
│   └── docker/             # Docker configurations
├── internal/
│   ├── config/             # Configuration handling
│   ├── docs/               # API documentation
│   ├── handler/            # HTTP handlers
│   ├── middleware/         # HTTP middleware
│   ├── migrations/         # Database migrations
│   ├── model/              # Data models
│   ├── repository/         # Data access layer
│   ├── router/             # Route definitions
│   └── service/            # Business logic
├── pkg/                    # Shared packages
│   ├── database/           # Database connection
│   └── utils/              # Utility functions
├── scripts/                # Utility scripts
└── test/                   # Test files
```

## Development

### Running Tests

```bash
make test
```

### Code Linting

```bash
make lint
```

### Building for Production

```bash
make build
```

## License

[MIT License](LICENSE)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request