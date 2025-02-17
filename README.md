
# Chirpy

Chirpy is a social media platform for sharing short messages called chirps. This project includes a backend server written in Go, a PostgreSQL database, and various API endpoints for managing users and chirps.

## Setup Instructions

### Prerequisites

- Go 1.16+
- PostgreSQL 12+

### Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/gaba-bouliva/Chirpy.git
    cd Chirpy
    ```

2. Create a `.env` file in the root directory with the following content:
    ```properties
    DB_URL="postgres://<username>:<password>@localhost:5432/chirpy?sslmode=disable"
    PLATFORM="dev"
    TOKEN_SECRET="<your_token_secret>"
    POLKA_KEY="<your_polka_key>"
    ```

3. Install dependencies:
    ```sh
    go mod tidy
    ```

4. Run the server:
    ```sh
    go run main.go
    ```

## Environment Variables

- `DB_URL`: The URL for connecting to the PostgreSQL database.
- `PLATFORM`: The environment in which the application is running (e.g., `dev`, `prod`).
- `TOKEN_SECRET`: The secret key used for signing JWT tokens.
- `POLKA_KEY`: The API key for Polka webhooks.

## API Endpoints

### User Endpoints

- `POST /api/users`: Create a new user.
- `POST /api/login`: Log in a user and return JWT and refresh tokens.
- `PUT /api/users`: Update user information.
- `POST /api/refresh`: Refresh the JWT token.
- `POST /api/revoke`: Revoke the refresh token.

### Chirp Endpoints

- `POST /api/chirps`: Create a new chirp.
- `GET /api/chirps`: Get all chirps.
- `GET /api/chirps/{id}`: Get a chirp by ID.
- `DELETE /api/chirps/{id}`: Delete a chirp by ID.

### Admin Endpoints

- `POST /admin/reset`: Reset the metrics (only available in `dev` environment).
- `GET /admin/metrics`: Get the current metrics.

### Webhook Endpoints

- `POST /api/polka/webhooks`: Handle Polka webhooks for user upgrades.

### Health Check

- `GET /api/healthz`: Health check endpoint.

## License

This project is licensed under the MIT License.
