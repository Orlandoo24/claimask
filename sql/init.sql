-- auto-generated definition
create table order_id
(
    id          int unsigned auto_increment
        primary key,
    order_id    bigint unsigned                     not null,
    address     varchar(255)                        not null,
    json        text                                not null,
    insert_time timestamp default CURRENT_TIMESTAMP null,
    update_time timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP
);





DELETE FROM order_id where id > 0;
ALTER TABLE order_id AUTO_INCREMENT = 1;