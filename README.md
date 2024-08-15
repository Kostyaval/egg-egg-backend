# Egg App Backend

`go run ./cmd/server`


## Environment variables

| Name                   | R | Description                                                                      | Default      |
|------------------------|---|----------------------------------------------------------------------------------|--------------|
| **RUNTIME**            | F | `development`, `production`                                                      | `production` |
| **MONGODB_URI**        | T | e.g. `mongodb+srv://<user>:<password>@<cluster-url>?retryWrites=true&w=majority` |              |
| **REDIS_URI**          | T | e.g. `redis://<user>:<pass>@<redis-url>/<db>`                                    |              |
| **TELEGRAM_TOKEN**     | T |                                                                                  |              |
| **JWT_PRIVATE_KEY**    | T |                                                                                  |              |
| **JWT_PUBLIC_KEY**     | T |                                                                                  |              |
| **JWT_ISS**            | F |                                                                                  | `egg.one`    |
| **CORS_ALLOW_ORIGINS** | F |                                                                                  |              |
| **CORS_MAX_AGE**       | F | In seconds, e.g. `3600`. To disable caching completely, pass negative value      | `0`          |
| **API_KEY**            | F | Secret access token for development purposes. Used in header `X-Api-Key`         |              |

## Generate JWT keys

- `jose jwk gen -i '{"alg": "ES256"}' -o ./private.jwk`
- `jose jwk pub -i ./private.jwk -o ./public.jwk`