create
    definer = root@`%` procedure checkifalive()
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