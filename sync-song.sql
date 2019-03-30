use syncsong;

drop table if exists Queue;
drop table if exists Lobby;
drop table if exists Track;

create table Track(
    uri varchar(100) primary key,
    name varchar(200) not null,
    artist varchar(200) not null
);

create table Lobby(
    id varchar(4) primary key,
    name varchar(100) not null,
    mode int(1) not null,
    genre varchar(100) not null,
    public bool not null,
    currentUri varchar(100),
    
    foreign key (currentUri) references Track(uri)
);

create table Queue(
    lobbyID varchar(4),
    trackURI varchar(100),
    rank int(3) not null,

    primary key (lobbyID, trackURI),
    foreign key (lobbyID) references Lobby(id),
    foreign key (trackURI) references Track(uri)
);

insert into Track values('id1', 'song1', 'artist name 1');
insert into Track values('id2', 'song2', 'artist name 2');

insert into Lobby values('ABCD', 'Has queue', 1, 'Rock', 1, 'id1');
insert into Lobby values('WXYZ', 'Has track', 2, 'pop', 0, 'id2');
insert into Lobby values('LKJS', 'Nothing', 2, 'pop', 0, null);

insert into Queue values('ABCD', 'id1', 1);
insert into Queue values('ABCD', 'id2', 2);

