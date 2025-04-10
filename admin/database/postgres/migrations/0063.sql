UPDATE usergroups SET name = 'autogroup:users' WHERE name = 'all-users';
UPDATE usergroups SET name = 'autogroup:members' WHERE name = 'all-members';
UPDATE usergroups SET name = 'autogroup:guests' WHERE name = 'all-guests';
