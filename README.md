# Sync Song Server

Server component of final year project.

Before setting up this server, docker must be installed.

A guide can be found at: https://docs.docker.com/install/

## Setup

Clone the repo and cd into it.

Create a network for the containers to communicate on:

`docker network create sync-song-network`

Run the following commands to setup the mysql container:

```
docker run --network=sync-song-network -e MYSQL_ROOT_PASSWORD=sspassword --name mysql -d mysql:5.7
docker cp sync-song.sql mysql:/sync-song.sql
docker exec mysql /bin/bash -c "mysql -u root -psspassword < sync-song.sql"
```

Build the syng-song container:

`docker build -t sync-song-server .`

Run the sync-song container:

`docker run -it -p 8080:8080 --network=sync-song-network --name sync-song sync-song-server:latest`

## Files contributed by me.

All files in this repo have been contributed by me.

Any code taken from online sources is referenced with a comment above it.
