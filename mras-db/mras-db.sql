drop
    database if exists mras;
SET
    GLOBAL event_scheduler = ON;
create
    database mras;
use mras;

create table permissions
(
    id      INT NOT NULL UNIQUE AUTO_INCREMENT,
    admin   BOOL DEFAULT FALSE,
    canedit BOOL DEFAULT FALSE,

);

create table user
(
    id            INT         NOT NULL UNIQUE AUTO_INCREMENT,
    username      VARCHAR(50) NOT NULL UNIQUE,
    password      VARCHAR(64) NOT NULL,
    created_ad    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    avatar_id     VARCHAR(100),
    perm_id       INT         NOT NULL,
    refresh_token TEXT,
    reset_code    TEXT        NOT NULL,
    FOREIGN KEY (perm_id) REFERENCES permissions (id),
    primary key (id)
);

create table usergroup
(
    id      INT          NOT NULL UNIQUE AUTO_INCREMENT,
    name    VARCHAR(100) NOT NULL,
    perm_id INT          NOT NULL,
    FOREIGN KEY (perm_id) REFERENCES permissions (id),
    PRIMARY KEY (id)
);

create table room
(
    id          INT          NOT NULL UNIQUE AUTO_INCREMENT,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    dim_height  NUMERIC,
    dim_width   NUMERIC,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    primary key (id)
);

create table speaker
(
    id            INT          NOT NULL UNIQUE AUTO_INCREMENT,
    name          VARCHAR(100) NOT NULL,
    description   TEXT,
    pos_x         NUMERIC,
    pos_y         NUMERIC,
    room_id       INT          NOT NULL,
    ip_address    VARCHAR(15)  NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_lifesign TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    alive         BOOL      DEFAULT TRUE,
    FOREIGN KEY (room_id) REFERENCES room (id),
    primary key (id)
);

create table speakergroup
(
    id   INT          NOT NULL UNIQUE AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    primary key (id)
);

create table speakergroupspeakers
(
    speakergroup_id INT NOT NULL,
    speaker_id      INT NOT NULL,
    foreign key (speakergroup_id) references speakergroup (id),
    foreign key (speaker_id) references speaker (id)
);



create table permrooms
(
    perm_id INT NOT NULL,
    room_id INT NOT NULL,
    foreign key (perm_id) references permissions (id),
    foreign key (room_id) references room (id)
);

create table permspeakers
(
    perm_id    INT NOT NULL,
    speaker_id INT NOT NULL,
    foreign key (perm_id) references permissions (id),
    foreign key (speaker_id) references speaker (id)
);

create table permspeakergroups
(
    perm_id         INT NOT NULL,
    speakergroup_id INT NOT NULL,
    foreign key (perm_id) references permissions (id),
    foreign key (speakergroup_id) references speakergroup (id)
);

drop procedure if exists checkifalive;
create procedure checkifalive()
begin
    declare
        speaker_id int;
    declare
        diff int;
    declare
        finished integer default 0;
    declare
        curId cursor for
            SELECT id
            from speaker;

    declare
        continue handler for not found set finished = 1;

    open curId;

    updAlive
    :
    loop
        FETCH curId into speaker_id;
        select TIMESTAMPDIFF(SQL_TSI_SECOND, (SELECT last_lifesign from speaker), CURRENT_TIMESTAMP)
        into diff;
        if
            diff >= 30 then
            update speaker
            set alive = false
            where id = speaker_id;
        else
            update speaker
            set alive = true
            where id = speaker_id;
        end if;

        if
            finished = 1 then
            LEAVE updAlive;
        end if;
    end loop
        updAlive;
    close curId;
end;

drop
    event if exists alivecheck;
create
    event alivecheck
    on schedule every 30 SECOND
    on completion PRESERVE
    do
    CALL checkifalive();
