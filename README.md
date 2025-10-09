# pypi server

A minimal, self-hosted Python package index server compatible with pip and uv.
Supports local and S3 storage backends, authentication via htpasswd, and is easy to deploy with Docker.

## Features

- Compatible with pip and uv
- Local filesystem or S3-compatible storage
- Basic authentication via htpasswd
- Simple HTML and legacy upload endpoints

## Configuration

Here's the full configuration file:

```yaml
log_level: info
htpasswd: ./htpasswd

server:
  host: 0.0.0.0
  port: 8080
  read_header_timeout_seconds: 10
  graceful_shutdown_timeout_seconds: 15
  enable_access_logger: true

storage:
  kind: local

  local:
    path: ./data

  s3:
    bucket: my-bucket
    prefix: my-prefix
    region: us-west-2
    endpoint: http://localhost:9000
    use_path_style: true
    access_key: myaccesskey
    secret_key: mysecretkey
```

Set the storage backend (`local` or `s3`) and authentication file as needed.

### Configuration fields

| Field                                 | Description                                      | Options/Example                | Default         |
|---------------------------------------|--------------------------------------------------|-------------------------------|-----------------|
| `log_level`                           | Logging verbosity                                | `debug`, `info`, `warn`, `error` | `info`          |
| `htpasswd`                            | Path to htpasswd file for authentication          | `./htpasswd`                  | `./htpasswd`    |
| `server.host`                         | Host address to bind the server                   | `0.0.0.0`                     | (empty)         |
| `server.port`                         | Port to run the server                            | `8080`                        | `3000`          |
| `server.read_header_timeout_seconds`  | Timeout for reading request headers (seconds)     | `10`                          | `5`             |
| `server.graceful_shutdown_timeout_seconds` | Timeout for graceful shutdown (seconds)      | `15`                          | `10`            |
| `server.enable_access_logger`         | Enable access logging                             | `true`, `false`               | `true`          |
| `storage.kind`                        | Storage backend type                              | `local`, `s3`                 | `local`         |
| `storage.local.path`                  | Path for local storage                            | `./data`                      | `./data`        |
| `storage.s3.bucket`                   | S3 bucket name                                   | `my-bucket`                   | (none)          |
| `storage.s3.prefix`                   | S3 key prefix (optional)                         | `my-prefix`                   | (none)          |
| `storage.s3.region`                   | S3 region                                        | `us-west-2`                   | (none)          |
| `storage.s3.endpoint`                 | S3-compatible endpoint URL                       | `http://localhost:9000`       | (none)          |
| `storage.s3.use_path_style`           | Use path-style addressing                        | `true`, `false`               | (none)          |
| `storage.s3.access_key`               | S3 access key                                    | `myaccesskey`                 | (none)          |
| `storage.s3.secret_key`               | S3 secret key                                    | `mysecretkey`                 | (none)          |

## Launch Instructions

### Using Docker

You can use the pre-built image from GitHub Container Registry:

```sh
docker pull ghcr.io/jeongukjae/pypi-server
```

Run the container:

```sh
docker run \
    -v $(pwd)/config-dev.yaml:/config.yaml \
    -v $(pwd)/htpasswd:/htpasswd \
    ghcr.io/jeongukjae/pypi-server \
    --config=/config.yaml
```

## Contributing

Contributions are welcome! Please follow these things:

- Ensure code is formatted (`make format`) and passes lint/tests (`make lint test`).

## License

MIT License. See [LICENSE](LICENSE) for details.
