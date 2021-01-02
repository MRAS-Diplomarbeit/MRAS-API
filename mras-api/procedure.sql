create
definer = root@`%` procedure checkifalive()
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
from speakers;

declare
continue handler for not found set finished = 1;

open curId;

updAlive
:
    loop
        FETCH curId into speaker_id;
select CURRENT_TIME - TIME_TO_SEC((SELECT last_lifesign from speakers))
into diff;
insert into difference(test) value (diff);
if
diff >= 30 then
update speakers
set alive = false
where id = speaker_id;
else
update speakers
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


