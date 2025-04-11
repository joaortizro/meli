# Architecture for a reverse proxy

## Features

* Acts as a reverse proxy for `https://api.mercadolibre.com/`.
* Using Docker for easy setup and consistent deployment.
* Redis for API rate limiter
* Logging

Follow these steps to get the reverse proxy running:

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/joaortizro/meli.git
    cd meli
    ```

2.  **Build and Run the container:**
    ```bash
    docker compose up --build
    ```
