CREATE TYPE gender_type AS ENUM ('U', 'M', 'F');
CREATE TYPE user_level_type AS ENUM ('Admin', 'User');

create table users(
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(100) NOT NULL,
    address varchar(255) NOT NULL,
    phone_number varchar(15) NOT NULL,
    gender gender_type NOT NULL,
    birth_date DATE NOT NULL,
    img_profile TEXT NOT NULL,
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp,
    deleted_at timestamptz 
);

INSERT INTO public.users
(id, "name", address, phone_number, "gender", birth_date, img_profile)
VALUES('c9a0a164-23e0-4d5c-9b06-7e149bb77954'::uuid, 'admin', '-', '-', 'M'::public.gender_type, '2001-01-01', '-');

create table auth(
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    username varchar(25) NOT NULL,
    password varchar(255) NOT NULL,
    email varchar(100) NOT NULL,
    level user_level_type NOT NULL,
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp,
    deleted_at timestamptz 
);

INSERT INTO public.auth
(id, user_id, username, "password", email, "level")
VALUES('459e4118-6d02-45d0-94a7-db5375dbc86d'::uuid, 'c9a0a164-23e0-4d5c-9b06-7e149bb77954'::uuid, 'admin', 'd576698f74fb8cdf2f7b738a9573995ccdd8fbe2551dd17d62fea9cfd1961c86', 'admin@epublib.co.id', 'Admin'::public."user_level_type");
