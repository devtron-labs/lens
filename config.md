# LENS CONFIGMAP

| Key                  | Value                                | Description                               |
|----------------------|--------------------------------------|-------------------------------------------|
| GIT_SENSOR_PROTOCOL  | GRPC                                 | The protocol used by the Git Sensor      |
| GIT_SENSOR_URL       | git-sensor-service.devtroncd:90       | The URL of the Git Sensor Service         |
| NATS_SERVER_HOST     | nats://devtron-nats.devtroncd:4222   | The host of the NATS server               |
| PG_ADDR              | postgresql-postgresql.devtroncd      | The address of the PostgreSQL server     |
| PG_DATABASE          | lens                                 | The name of the PostgreSQL database       |
| PG_PORT              | "5432"                               | The port number for PostgreSQL            |
| PG_USER              | postgres                             | The username for PostgreSQL access       |
