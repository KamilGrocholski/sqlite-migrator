alter table user
add role_id integer references user_role(id);
