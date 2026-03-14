#!/bin/bash
git pull origin master
docker compose build backend
docker compose up -d backend