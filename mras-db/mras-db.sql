create table permissions
(
	id int auto_increment,
	admin tinyint(1) default 0 not null,
	can_edit tinyint(1) default 0 not null,
	constraint id
		unique (id)
);

alter table permissions
	add primary key (id);

create table rooms
(
	id int auto_increment,
	name varchar(100) not null,
	description longtext null,
	dim_height double null,
	dim_width double null,
	created_at datetime(3) null,
	height double null,
	width double null,
	constraint id
		unique (id)
);

alter table rooms
	add primary key (id);

create table perm_rooms
(
	permissions_id int not null,
	room_id int not null,
	primary key (permissions_id, room_id),
	constraint fk_perm_rooms_permissions
		foreign key (permissions_id) references permissions (id)
			on update cascade on delete cascade,
	constraint fk_perm_rooms_room
		foreign key (room_id) references rooms (id)
			on update cascade on delete cascade
);

create table speaker_groups
(
	id int auto_increment,
	name varchar(100) not null,
	constraint id
		unique (id)
);

alter table speaker_groups
	add primary key (id);

create table perm_speakergroups
(
	permissions_id int not null,
	speaker_group_id int not null,
	primary key (permissions_id, speaker_group_id),
	constraint fk_perm_speakergroups_permissions
		foreign key (permissions_id) references permissions (id)
			on update cascade on delete cascade,
	constraint fk_perm_speakergroups_speaker_group
		foreign key (speaker_group_id) references speaker_groups (id)
			on update cascade on delete cascade
);

create table speakers
(
	id int auto_increment,
	name varchar(100) null,
	description longtext null,
	pos_x double null,
	pos_y double null,
	room_id int null,
	ip_address longtext null,
	created_at datetime(3) null,
	last_lifesign datetime(3) null,
	alive tinyint(1) default 1 not null,
	constraint id
		unique (id)
);

alter table speakers
	add primary key (id);

create table perm_speakers
(
	permissions_id int not null,
	speaker_id int not null,
	primary key (permissions_id, speaker_id),
	constraint fk_perm_speakers_permissions
		foreign key (permissions_id) references permissions (id)
			on update cascade on delete cascade,
	constraint fk_perm_speakers_speaker
		foreign key (speaker_id) references speakers (id)
			on update cascade on delete cascade
);

create table sessions
(
	id int auto_increment,
	speaker_id int null,
	display_name longtext null,
	method longtext null,
	created_at datetime(3) null,
	constraint id
		unique (id),
	constraint fk_sessions_speaker
		foreign key (speaker_id) references speakers (id)
);

alter table sessions
	add primary key (id);

create table session_speakers
(
	sessions_id int not null,
	speaker_id int not null,
	primary key (sessions_id, speaker_id),
	constraint fk_session_speakers_sessions
		foreign key (sessions_id) references sessions (id)
			on update cascade on delete cascade,
	constraint fk_session_speakers_speaker
		foreign key (speaker_id) references speakers (id)
			on update cascade on delete cascade
);

create table speakergroup_speakers
(
	speaker_group_id int not null,
	speaker_id int not null,
	primary key (speaker_group_id, speaker_id),
	constraint fk_speakergroup_speakers_speaker
		foreign key (speaker_id) references speakers (id)
			on update cascade on delete cascade,
	constraint fk_speakergroup_speakers_speaker_group
		foreign key (speaker_group_id) references speaker_groups (id)
			on update cascade on delete cascade
);

create table user_groups
(
	id int auto_increment,
	name varchar(100) null,
	perm_id int null,
	constraint id
		unique (id),
	constraint fk_user_groups_permissions
		foreign key (perm_id) references permissions (id)
);

alter table user_groups
	add primary key (id);

create table users
(
	id int auto_increment,
	username varchar(15) null,
	password varchar(64) null,
	created_at datetime(3) null,
	avatar_id varchar(10) default 'default' null,
	perm_id int null,
	refresh_token longtext null,
	reset_code longtext null,
	password_reset tinyint(1) default 0 null,
	constraint id
		unique (id),
	constraint username
		unique (username),
	constraint fk_users_permissions
		foreign key (perm_id) references permissions (id)
);

alter table users
	add primary key (id);

create table user_usergroups
(
	user_group_id int not null,
	user_id int not null,
	primary key (user_group_id, user_id),
	constraint fk_user_usergroups_user
		foreign key (user_id) references users (id)
			on update cascade on delete cascade,
	constraint fk_user_usergroups_user_group
		foreign key (user_group_id) references user_groups (id)
			on update cascade on delete cascade
);

create view room_user_perms as
	select `mras`.`rooms`.`id` AS `room_id`, `mras`.`users`.`id` AS `user_id`
from (`mras`.`users`
         join `mras`.`rooms` on ((`mras`.`rooms`.`id` in (select `mras`.`perm_rooms`.`room_id`
                                                          from `mras`.`perm_rooms`
                                                          where ((`mras`.`perm_rooms`.`permissions_id` = `mras`.`users`.`perm_id`) or
                                                                 `mras`.`perm_rooms`.`permissions_id` in
                                                                 (select `mras`.`user_usergroups_perms`.`perm_id`
                                                                  from `mras`.`user_usergroups_perms`
                                                                  where (`mras`.`user_usergroups_perms`.`user_id` = `mras`.`users`.`id`)))) or
                                  (0 <> (select `mras`.`permissions`.`admin`
                                         from `mras`.`permissions`
                                         where ((`mras`.`permissions`.`id` = `mras`.`users`.`perm_id`) = true))) or
                                  (0 <> (select `mras`.`permissions`.`admin`
                                         from `mras`.`permissions`
                                         where (`mras`.`permissions`.`id` in
                                                (select `mras`.`user_usergroups_perms`.`perm_id`
                                                 from `mras`.`user_usergroups_perms`
                                                 where (`mras`.`user_usergroups_perms`.`user_id` = `mras`.`users`.`id`)) =
                                                true))))))
group by `room_id`, `user_id`;

create view speaker_user_perms as
	select `mras`.`speakers`.`id` AS `speaker_id`, `mras`.`users`.`id` AS `user_id`
from (`mras`.`users`
         join `mras`.`speakers` on ((`mras`.`speakers`.`id` in (select `mras`.`perm_speakers`.`speaker_id`
                                                                from `mras`.`perm_speakers`
                                                                where ((`mras`.`perm_speakers`.`permissions_id` = `mras`.`users`.`perm_id`) or
                                                                       `mras`.`perm_speakers`.`permissions_id` in
                                                                       (select `mras`.`user_usergroups_perms`.`perm_id`
                                                                        from `mras`.`user_usergroups_perms`
                                                                        where (`mras`.`user_usergroups_perms`.`user_id` = `mras`.`users`.`id`)))) or
                                     (0 <> (select `mras`.`permissions`.`admin`
                                            from `mras`.`permissions`
                                            where ((`mras`.`permissions`.`id` = `mras`.`users`.`perm_id`) = true))) or
                                     (0 <> (select `mras`.`permissions`.`admin`
                                            from `mras`.`permissions`
                                            where (`mras`.`permissions`.`id` in
                                                   (select `mras`.`user_usergroups_perms`.`perm_id`
                                                    from `mras`.`user_usergroups_perms`
                                                    where (`mras`.`user_usergroups_perms`.`user_id` = `mras`.`users`.`id`)) =
                                                   true))))))
group by `speaker_id`, `user_id`;

create view speakergroup_user_perms as
	select `mras`.`speaker_groups`.`id` AS `speakergoup_id`, `mras`.`users`.`id` AS `user_id`
from (`mras`.`users`
         join `mras`.`speaker_groups`
              on ((`mras`.`speaker_groups`.`id` in (select `mras`.`perm_speakergroups`.`speaker_group_id`
                                                    from `mras`.`perm_speakergroups`
                                                    where ((`mras`.`perm_speakergroups`.`permissions_id` =
                                                            `mras`.`users`.`perm_id`) or
                                                           `mras`.`perm_speakergroups`.`permissions_id` in
                                                           (select `mras`.`user_usergroups_perms`.`perm_id`
                                                            from `mras`.`user_usergroups_perms`
                                                            where (`mras`.`user_usergroups_perms`.`user_id` = `mras`.`users`.`id`)))) or
                   (0 <> (select `mras`.`permissions`.`admin`
                          from `mras`.`permissions`
                          where ((`mras`.`permissions`.`id` = `mras`.`users`.`perm_id`) = true))) or
                   (0 <> (select `mras`.`permissions`.`admin`
                          from `mras`.`permissions`
                          where (`mras`.`permissions`.`id` in (select `mras`.`user_usergroups_perms`.`perm_id`
                                                               from `mras`.`user_usergroups_perms`
                                                               where (`mras`.`user_usergroups_perms`.`user_id` = `mras`.`users`.`id`)) and
                                 (`mras`.`permissions`.`admin` = true)))))))
group by `speakergoup_id`, `user_id`;

create view user_usergroups_perms as
	select `mras`.`users`.`id` AS `user_id`, `mras`.`user_groups`.`perm_id` AS `perm_id`
from (`mras`.`users`
         join `mras`.`user_groups` on (`mras`.`user_groups`.`id` in (select `mras`.`user_usergroups`.`user_group_id`
                                                                     from `mras`.`user_usergroups`
                                                                     where (`mras`.`user_usergroups`.`user_id` = `mras`.`users`.`id`))))
group by `mras`.`users`.`id`, `mras`.`user_groups`.`perm_id`;

create procedure checkifalive()
begin
    declare speaker_id int; declare diff int; declare finished integer default 0;
    declare curId cursor for SELECT id from speakers; declare continue handler for not found set finished = 1;
    open curId;
    updAlive:
    loop
        FETCH curId into speaker_id;
        select TIMESTAMPDIFF(SECOND, (SELECT last_lifesign from speakers where id = speaker_id), CURRENT_TIMESTAMP)
        into diff;
        if diff >= 30 then
            update speakers set alive = 0 where id = speaker_id;
        else
            update speakers set alive = 1 where id = speaker_id;
        end if;
        if finished = 1 then LEAVE updAlive; end if;
    end loop updAlive;
    close curId;
end;

create event alivecheck on schedule
	every '30' SECOND
	starts '2021-03-18 13:44:05'
	on completion preserve
	enable
	do
	CALL checkifalive();
