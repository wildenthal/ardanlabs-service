# Ardan Labs Service

This is code corresponding to the "Software Design with Kubernetes 2.0" course from Ardan Labs.
The course is available at [https://www.ardanlabs.com/](https://www.ardanlabs.com/).

## Description

Currently includes a single service set up to run in a Kubernetes cluster using Kind.
The original code is from the course, with some modifications: 
 - Podman is used instead of Docker
 - OTEL is used for tracing and visualized with Jaeger
 - Simplified abstractions where direct approaches felt more appropriate

## Getting Started

### Dependencies

* Go 1.24
* Podman
* Kind
* Kubectl

### Executing program

See the makefile for the available commands.

## Environment Variables

The following environment variables are used to configure the application:

- `READ_TIMEOUT`: The maximum duration for reading the entire request, including the body. Default is `5s`.
- `WRITE_TIMEOUT`: The maximum duration before timing out writes of the response. Default is `10s`.
- `IDLE_TIMEOUT`: The maximum amount of time to wait for the next request when keep-alives are enabled. Default is `120s`.
- `SHUTDOWN_TIMEOUT`: The maximum amount of time to wait for the server to shut down gracefully. Default is `5s`.
- `OTEL_EXPORTER_OTLP_ENDPOINT`: The endpoint for the OpenTelemetry collector. This is a required variable.
- `API_HOST`: The host and port for the API server. Default is `0.0.0.0:3000`.
- `DEBUG_HOST`: The host and port for the debug server. Default is `0.0.0.0:3010`.

## Acknowledgments

* [ardanlabs](https://github.com/ardanlabs/)
