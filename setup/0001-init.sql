create table if not exists urls (
    id int primary key,
    url varchar(255),
    description varchar(255),
    queue varchar(255),
    domain varchar(255),
    is_active boolean,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp          
)

