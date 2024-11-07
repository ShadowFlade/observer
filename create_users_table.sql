create table stats (id int auto_increment not null, mem_usage float not null, mem_usage_percent float not null ,date_inserted datetime not null, primary key (`id`), user_id int not null, day_active_users int not null);

