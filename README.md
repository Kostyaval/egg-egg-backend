# Egg App Backend

`go run ./cmd/server`


## Environment variables

| Name                     | R | Description                                                                      | Default      |
|--------------------------|---|----------------------------------------------------------------------------------|--------------|
| **RUNTIME**              | F | `development`, `production`                                                      | `production` |
| **MONGODB_URI**          | T | e.g. `mongodb+srv://<user>:<password>@<cluster-url>?retryWrites=true&w=majority` |              |
| **REDIS_URI**            | T | e.g. `redis://<user>:<pass>@<redis-url>/<db>`                                    |              |
| **TELEGRAM_TOKEN**       | T |                                                                                  |              |
| **JWT_PRIVATE_KEY_PATH** | T |                                                                                  |              |
| **JWT_PUBLIC_KEY_PATH**  | T |                                                                                  |              |
| **JWT_ISS**              | F |                                                                                  | `egg.one`    |
| **JWT_TTL**              | F |                                                                                  | `15m`        |
| **RULES_PATH**           | T | Path to yaml file described the game rules                                       |              |


## Generate JWT keys

- `jose jwk gen -i '{"alg": "ES256"}' -o ./private.jwk`
- `jose jwk pub -i ./private.jwk -o ./public.jwk`