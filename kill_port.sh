#!/bin/bash
PORT=8080
kill -9 $(lsof -t -i:${PORT})