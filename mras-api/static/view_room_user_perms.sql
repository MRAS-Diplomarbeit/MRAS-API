create or replace view room_user_perms as
select rooms.id as room_id, users.id as user_id
from users
         inner join rooms on ((rooms.id = any (select perm_rooms.room_id
                                               from perm_rooms
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
group by room_id, user_id;