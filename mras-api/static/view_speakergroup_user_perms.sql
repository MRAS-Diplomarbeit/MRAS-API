create or replace view speakergroup_user_perms as
select speaker_groups.id as speakergoup_id, users.id as user_id
from users
         inner join speaker_groups on ((speaker_groups.id = any (select perm_speakergroups.speaker_group_id
                                                                 from perm_speakergroups
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
group by speakergoup_id, user_id;