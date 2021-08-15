#!/bin/bash

go test -v -coverprofile=cover.out ./...
go tool cover -html=cover.out -o coverage.html