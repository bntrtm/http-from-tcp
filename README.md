# http-from-tcp
A light implementation of the HTTP 1.1 protocol built on TCP sockets.

## Motivation

I built this project to follow alongside an advanced [Boot.Dev](https://www.boot.dev/courses/learn-http-protocol-golang) course, which had the objective of demystifying the core logic of the HTTP protocol by having me implement it in Go myself.

## Features

This project provides types under the `./internal` packages which may be used to parse incoming byte streams into structured HTTP Requests, as well as serialize HTTP Responses.

An example of the library in action can be found under `./cmd/httpserver`.

