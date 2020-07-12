create table if not exists users (
	id serial primary key,
	username varchar(255),
	userID integer default 0,
	blacklist boolean default false,
	isAdmin boolean default false,
	coins integer default 100,
	reputation integer default 100,
	weapons_power integer default 0,
	activ_date DATE default CURRENT_TIMESTAMP,
	activity int default 10
);

insert into users (username, isAdmin, coins, userID)
select 'likipiki', true, 1000, 216399855
where not exists (
	select 1 from users where username='likipiki'
);

create table if not exists weapons (
	id serial primary key,
	name varchar(50),
	power int default 0,
	cost int default 0
);

insert into weapons (name, power, cost) values('Палка', 1, 10);
insert into weapons (name, power, cost) values('Копыто', 5, 40);
insert into weapons (name, power, cost) values('Водный пистолет', 6, 50);
insert into weapons (name, power, cost) values('Ленивый дробовик', 14, 100);
insert into weapons (name, power, cost) values('БаЗЗЗука', 32, 150);
insert into weapons (name, power, cost) values('Катапульта', 111, 450);
insert into weapons (name, power, cost) values('Зубодробящий крутокрякс', 111, 450);
insert into weapons (name, power, cost) values('Попапылающий огнемет', 115, 470);
insert into weapons (name, power, cost) values('Огурец 🥒', 50, 2000);
