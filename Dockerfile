FROM golang:1.11
#FROM mysql:8

WORKDIR /go/src/github.com/Red350/sync-song-server

COPY . .

ENV MYSQL_ALLOW_EMPTY_PASSWORD true

RUN apt-get update
RUN apt-get -y install mysql-server

#RUN mysql -u root < ./sync-song.sql 
#RUN /etc/init.d/mysql start && \
#RUN /usr/sbin/mysqld start 
#&& \
#    mysql -u root < ./sync-song.sql 
#    #mysql -h localhost -P 3306 --protocol=tcp -u root < ./sync-song.sql 
#RUN apt-get -y install golang 
RUN go get -d -v .
RUN go install -v .
#RUN mysql -u root < ./sync-song.sql 

CMD mysql -h 127.0.0.1 -P 3306 --protocol=tcp -u root < ./sync-song.sql && /go/bin/sync-song-server

EXPOSE 8080

