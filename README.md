# Verancial

Verancial is a service that helps users manage their financial information. It allows users to securely download, store, and process their financial information from banks. This information can be used for data analysis and exporting reports to popular financial apps like BudgetBakers and YNAB.

## Installation

To run Verancial, you'll need to have the following tools installed:

- Docker
- Go 1.18

Once you have these installed, follow these steps:

1. Clone the repository:

    ```
    git clone https://github.com/verasthiago/verancial.git
    ```

2. Navigate to the project root directory:

    ```
    cd verancial
    ```

3. Run the database migrations:

    ```
    make migrate_db
    ```

4. Build and run the services using the Makefile:

    ```
    make all_build
    ```

5. Access the application:
    - Frontend: http://localhost:3000
    - API: http://localhost:8080
    - Login: http://localhost:8081

## Services

Verancial is a monorepo consisting of the following services:

### Frontend

The Frontend service is a React TypeScript application that provides a modern web interface for users to:
- Login securely with JWT authentication
- View dashboard with bank account statistics
- See transaction counts and data freshness for each connected bank
- Manage bank account connections

### API

The API service provides an interface for accessing and managing user financial data, including:
- Dashboard statistics endpoints
- Bank account management
- Transaction data processing

### Login

The Login service provides user authentication and authorization functionality.

### Shared

The Shared folder contains Go code that is shared across all services.

## Contributing

We welcome contributions to Verancial. To contribute:

1. Open an issue to discuss your changes.
2. Fork the repository.
3. Create a new branch for your changes.
4. Make your changes and commit them.
5. Push your changes to your forked repository.
6. Open a pull request to the main repository.

## Contact

If you have any questions or concerns, you can reach us at `verancial@verasthiago.com` or on LinkedIn at https://www.linkedin.com/in/verasthiago/.

## Future Plans

We plan to add a CLI in the near future, and possibly a frontend. We're also working on a new service that's currently in development. Stay tuned for more updates!
