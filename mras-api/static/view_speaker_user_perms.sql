create or replace view speaker_user_perms as
select speakers.id as speaker_id, users.id as user_id
from users
         inner join speakers on ((speakers.id = any (select perm_speakers.speaker_id
                                                     from perm_speakers
                                                     where permissions_id = users.perm_id
                                                        or permissions_id = any
                                                           (select perm_id from user_usergroups_perms where user_id = users.id)))
    or (select admin
        from permissions
        where id = users.perm_id = true)
    or (select admin
        from permissions
        where permissions.id = any
              (select perm_id from user_usergroups_perms where user_id = users.id)
          and admin = true))
group by speaker_id, user_id;