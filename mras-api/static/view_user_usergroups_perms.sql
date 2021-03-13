create or replace view user_usergroups_perms as
select users.id as user_id, user_groups.perm_id as perm_id
from users
         inner join user_groups
                    on (user_groups.id = any (select user_group_id from user_usergroups where user_id = users.id))
group by users.id, user_groups.perm_id;