create database if not exists syncsong;
use syncsong;

drop table if exists queue;
drop table if exists lobby;
drop table if exists track;

create table track(
    uri varchar(100) primary key,
    name varchar(200) not null,
    artist varchar(200) not null,
    duration bigint not null
);

create table lobby(
    id varchar(4) primary key,
    name varchar(100) not null,
    mode int(1) not null,
    genre varchar(100) not null,
    public bool not null,
    currentUri varchar(100),
    
    foreign key (currentUri) references track(uri)
);

create table queue(
    lobbyID varchar(4),
    trackURI varchar(100),
    _rank int(3) not null,

    primary key (lobbyID, trackURI),
    foreign key (lobbyID) references lobby(id),
    foreign key (trackURI) references track(uri)
);

# Test data.
#insert into track values('id1', 'song1', 'artist name 1');
#insert into track values('id2', 'song2', 'artist name 2');
#
#insert into lobby values('ABCD', 'Has queue', 1, 'Rock', 1, 'id1');
#insert into lobby values('WXYZ', 'Has track', 2, 'pop', 0, 'id2');
#insert into lobby values('LKJS', 'Nothing', 2, 'pop', 0, null);
#
#insert into queue values('ABCD', 'id1', 1);
#insert into queue values('ABCD', 'id2', 2);

