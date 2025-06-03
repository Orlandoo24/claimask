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


-- auto-generated definition
create table rich_reward_log
(
    id            bigint auto_increment comment '主键 id'
        primary key,
    address       varchar(64)                             not null comment '用户钱包地址',
    total_reward  decimal(9, 1) default 0.0               not null comment '累加收益，领取后置0',
    update_reward date                                    null comment '累加收益最新时间',
    latest        date                                    null comment '最后领取日期',
    latest_detail timestamp                               null comment '最后领取日期具体时间',
    status        bit           default b'0'              not null comment '今日是否已经领取，0未领取，1已领取',
    insert_time   timestamp     default CURRENT_TIMESTAMP not null comment '创建时间',
    update_time   timestamp     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    is_deleted    bit           default b'0'              not null comment '是否删除，0：正常，1：删除',
    constraint uk_address
        unique (address)
)
    comment '会员收益表';

