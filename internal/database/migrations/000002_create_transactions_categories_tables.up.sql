CREATE TABLE categories
(
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE transactions
(
    id          SERIAL PRIMARY KEY ,
    title       TEXT        NOT NULL,
    price       DECIMAL     NOT NULL,
    date_made   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    owner_id    UUID,
    category_id INT,
    type        VARCHAR(10),

    FOREIGN KEY (owner_id) REFERENCES users (id),
    FOREIGN KEY (category_id) REFERENCES categories (id),
    CHECK (type in ('Expense', 'Income'))

);