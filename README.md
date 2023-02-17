# fhir-to-server
> Load FHIR🔥 bundles from a Kafka topic into a FHIR server

## Concurrency

In order to enable Multi-threaded message consumption **one consumer per input topic** is used. 
Multiple consumers per topic are currently not supported.

## Offset handling

By default, the consumers are configured to auto-commit offsets, in order to improve performance.

However, the latest successfully processed messages (i.e. send to the FHIR server) per topic are 
committed manually on shutdown (interrupt or kill).
This ensures that offsets reflect successfully processed messages only. 

## Retry capabilities

The HTTP client supports retrying requests to the FHIR server in case the target endpoint is unavailable
or runs into a timeout. See [configuration properties](#configuration-properties) below.

## Validation

FHIR resource types are currently not validated. Processing requires only valid JSON content.

## Configuration properties

| Name                             | Default                    | Description                             |
|----------------------------------|----------------------------|-----------------------------------------|
| `app.name`                       | fhir-to-server             | Kafka consumer group id                 |
| `log-level`                      | info                       | Log level (error,warn,info,debug,trace) |
| `kafka.bootstrap-servers`        | localhost:9092             | Kafka brokers                           |
| `kafka.security-protocol`        | ssl                        | Kafka communication protocol            |
| `kafka.input-topic`              |                            | Kafka topic to consume                  |
| `kafka.ssl.ca-location`          | /app/cert/kafka-ca.pem     | Kafka CA certificate location           |
| `kafka.ssl.certificate-location` | /app/cert/app-cert.pem     | Client certificate location             |
| `kafka.ssl.key-location`         | /app/cert/app-key.pem      | Client  key location                    |
| `kafka.ssl.key-password`         | private-key-password       | Client key password                     |
| `fhir.server.base-url`           | http://localhost:8080/fhir | FHIR server base URL                    |
| `fhir.server.auth.user`          |                            | FHIR server BasicAuth username          |
| `fhir.server.auth.password`      |                            | FHIR server BasicAuth password          |
| `fhir.retry.count`               | 10                         | Retry count                             |
| `fhir.retry.timeout`             | 10                         | Retry timeout                           |
| `fhir.retry.wait`                | 5                          | Retry wait between retries              |
| `fhir.retry.max-wait`            | 20                         | Retry maximum wait                      |

### Environment variables

Override configuration properties by providing environment variables with their respective names.
Upper case env variables are supported as well as underscores (`_`) instead of `.` and `-`. 