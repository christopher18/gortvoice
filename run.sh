#!/bin/bash
# Load environment variables from .env file
export $(cat .env | xargs)

# Run the realtime application
go run cmd/realtime/*.go

