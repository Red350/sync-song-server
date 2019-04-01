# sync-song-server

Sync Song server

## Building and running using docker

Clone the repo, and execute the following commands:

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
