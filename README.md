# Sync Song Server

## Building and running using docker

Clone the repo, and execute the following commands to setup the mysql container:

```
docker run --name mysql -e MYSQL_ALLOW_EMPTY_PASSWORD=true -d mysql:latest -p 3306:3306
docker cp sync-song.sql mysql:/sync-song.sql
docker exec -it mysql /bin/bash
mysql -u root < sync-song.sql
```

Then execute these commands to build and run the sync-song container:

`docker build -t sync-song-server .`

`docker run -it -p 8080:8080 --rm --name sync-song sync-song-server:latest`

## GCP dockerless setup

Add a firewall rule for port 8080 ingress.

```
sudo apt-get upgrade
sudo apt-get install mysql-server
sudo mysql -u root < ./sync-song.sql
sudo apt-get install golang
go get -d -v .
go build -v .
./sync-song-server
```

