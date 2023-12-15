create table orders
(
    order_uid          varchar primary key,
    track_number       varchar,
    entry              varchar,
    delivery_id        bigint references deliveries (id),
    transaction        varchar references payments (transaction),
    locale             varchar,
    internal_signature varchar,
    customer_id        varchar,
    delivery_service   varchar,
    shardkey           varchar,
    sm_id              bigint,
    date_created       timestamp,
    oof_shard          varchar
);

create table deliveries
(
    id      serial primary key,
    name    varchar,
    phone   varchar(11),
    zip     varchar,
    city    varchar,
    address varchar,
    region  varchar,
    email   varchar
);

create table payments
(
    transaction   varchar primary key,
    request_id    varchar,
    currency      char(3),
    provider      varchar,
    amount        numeric(10, 2),
    payment_dt    bigint,
    bank          varchar,
    delivery_cost numeric(10, 2),
    goods_total   numeric(10, 2),
    custom_fee    numeric(10, 2)
);

create table items
(
    chrt_id      bigint primary key,
    track_number varchar,
    price        numeric(10, 2),
    rid          varchar,
    name         varchar,
    sale         numeric(10, 2),
    size         varchar,
    total_price  numeric(10, 2),
    nm_id        bigint,
    brand        varchar,
    status       int,
    order_uid    varchar references orders (order_uid)
);
