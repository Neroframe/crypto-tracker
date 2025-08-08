# crypto-tracker

Simple crypto price tracker.
Cryptocurrencies list available at `https://coinpaprika.com/`

## How to Start

Run: `docker-compose up --build`

This will:

- Start PostgreSQL

- Apply DB migrations from ./migrations

- Launch the API at http://localhost:8080

## Available API Routes

- `POST /currency/add` — Add a cryptocurrency to the tracking list
- `POST /currency/remove` — Remove a cryptocurrency from the tracking list
- `POST /currency/price` — Get the price of a cryptocurrency at a specific timestamp (returns the closest price ≤ timestamp)

## API Docs

Swagger UI is available at:  
`http://localhost:8080/swagger/index.html`
