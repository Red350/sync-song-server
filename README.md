# sync-song-server

Sync Song server

## Building and running using docker

Clone the repo, cd into its directory, and execute the following commands:

`docker run --name ss-mysql -e MYSQL_ROOT_PASSWORD=secret -d mysql:5.6`

`docker exec -i ss-mysql mysql -u root -psecret < ./sync-song.sql`

`docker build -t sync-song-server .`

`docker run -it -p 8080:8080 --rm --name sync-song sync-song-server:latest`

