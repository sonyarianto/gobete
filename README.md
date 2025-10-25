# gobete

Hi! gobete is short for "Go backend template". It is a boilerplate backend API built with Go and the Fiber web framework. It provides a solid foundation for building RESTful APIs.

## Features

- User authentication and management.
- Connection to MySQL database. Using GORM as the ORM layer.
- Easy to understand and extend, as long as you follow the existing structure and understand Go and Fiber basics. Everything starts from `main.go`. The `internal` directory contains the core application logic, organized into subdirectories for different modules and functionalities.
- Environment configuration using `.env` file.

## Goals
- Provide a robust and scalable backend for any web application.
- Ensure security and efficiency in handling user data and requests.
- Facilitate easy integration with frontend applications and other services.
- Maintain clean and maintainable codebase for future development.
- Follow best practices in Go programming and web development.
- Support OpenAPI/Swagger documentation for easy API exploration and testing.
- Implement comprehensive error handling and logging.
- Use middleware for tasks like logging, CORS, and authentication.
- Enable hot reloading during development for faster iteration.
- Support deployment in various environments, including local, staging, and production.
- Unit and integration testing to ensure code quality and reliability.

## Getting started

To get started with the gobete API, follow these steps:

1. Clone the repository:
   ```bash
   git clone https://github.com/sonyarianto/gobete.git
   ```

2. Navigate to the project directory:
   ```bash
   cd gobete
   ```

3. Install the required dependencies:
   ```bash
   go mod tidy
   ```

4. Set up your environment variables:
   Copy the `.env.example` file to `.env` and fill in the required values.

5. Run the application:
   ```bash
   go run main.go
   ```
6. The API will be available at `http://localhost:9000` (or the port you specified in the `.env` file).

## Hot reload during development
For development, you can use `air` for hot reloading. Install it using:
```bash
go install github.com/air-verse/air@latest
```
Then run:
```bash
air
```

## Git hooks

This project using `lefthook` for Git hooks. Follow the instructions at [lefthook installation guide](https://lefthook.dev/installation/go.html).

Run the following command to install the Git hooks defined in this project:
```bash
lefthook install
```

This will set up the pre-commit hook to run commands before each commit. See on `lefthook.yml` for more details.

## Some planned features and TODOs
- Add unit and integration tests.
- Implement more advanced features like role-based access control.
- Audit logging and monitoring.
- Better documentation and examples.
- Docker support for easier deployment.

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## Author
- Sony AK - [sony@sony-ak.com](https://sony-ak.com)