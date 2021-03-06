#!/bin/bash
cd "$(dirname $0)"

go test -coverprofile=cover.out commands/*
go tool cover -html=cover.out -o cover.html
open cover.html
