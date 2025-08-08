# crypto-tracker

Simple crypto price tracker using the Coinpaprika API.
Cryptocoin symbols must match those listed on [`https://coinpaprika.com/`](https://coinpaprika.com/)

## How to Start

Run: `docker-compose up --build`

This will:

- Start PostgreSQ
- Apply DB migrations from `./migrations`
- Launch the API at `http://localhost:8080`

## Available API Routes

- `POST /currency/add` — Add a cryptocurrency to the tracking list
- `POST /currency/remove` — Remove a cryptocurrency from the tracking list
- `POST /currency/price` — Get the price of a cryptocurrency at a specific timestamp (returns the closest nearest price in USD)

> **Note:** The `./config/dev.yaml` config file sets the fetch interval to **30 seconds**.
> You can change it, but keep in mind coinpaprika rate limits.
>
> ```yaml
> external:
>   coinpaprikaUrl: "https://api.coinpaprika.com"
>   rateLimit: 10.0
>   fetchInterval: "30s"
> ```

## API Docs

Swagger UI is available at:  
`http://localhost:8080/swagger/index.html`
