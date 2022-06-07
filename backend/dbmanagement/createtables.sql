
CREATE TABLE users (
    userid serial primary key,
    name varchar(90),
    username varchar(90) unique,
    password varchar(90),
    email varchar(90) unique
);



