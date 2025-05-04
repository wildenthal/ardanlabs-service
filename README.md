# Ardan Labs Service

This is code corresponding to the "Software Design with Kubernetes 2.0" course from Ardan Labs.
The course is available at [https://www.ardanlabs.com/](https://www.ardanlabs.com/).

## Description

Currently includes a single service set up to run in a Kubernetes cluster using Kind.
The original code is from the course, with some modifications: 
 - Podman is used instead of Docker
 - OTEL is used for tracing
 - Dependencies are simplified in favor of the standard library

## Getting Started

### Dependencies

* Go 1.24
* Podman
* Kind
* Kubectl

### Executing program

See the makefile for the available commands.

## Acknowledgments

* [ardanlabs](https://github.com/ardanlabs/)
