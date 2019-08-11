#!/usr/bin/env bash
git pull
docker-compose build media
docker-compose down
docker-compose up -d