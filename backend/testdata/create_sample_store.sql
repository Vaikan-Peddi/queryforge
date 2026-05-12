PRAGMA foreign_keys = ON;

DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS customers;

CREATE TABLE customers (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    city TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE products (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    price REAL NOT NULL
);

CREATE TABLE orders (
    id INTEGER PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES customers(id),
    order_date TEXT NOT NULL,
    status TEXT NOT NULL
);

CREATE TABLE order_items (
    id INTEGER PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price REAL NOT NULL,
    line_total REAL NOT NULL,
    FOREIGN KEY(order_id) REFERENCES orders(id),
    FOREIGN KEY(product_id) REFERENCES products(id)
);

INSERT INTO customers (id, name, email, city, created_at) VALUES
(1, 'Alice Johnson', 'alice@example.com', 'New York', '2025-01-04'),
(2, 'Bruno Kim', 'bruno@example.com', 'Seattle', '2025-01-09'),
(3, 'Carla Singh', 'carla@example.com', 'Austin', '2025-02-12'),
(4, 'Diego Torres', 'diego@example.com', 'Miami', '2025-02-20');

INSERT INTO products (id, name, category, price) VALUES
(1, 'Analytics Notebook', 'Software', 49.00),
(2, 'Data Studio Seat', 'Software', 129.00),
(3, 'SQL Field Guide', 'Books', 24.00),
(4, 'Query Credits Pack', 'Usage', 75.00);

INSERT INTO orders (id, customer_id, order_date, status) VALUES
(1, 1, '2025-03-01', 'paid'),
(2, 2, '2025-03-03', 'paid'),
(3, 1, '2025-03-09', 'paid'),
(4, 3, '2025-03-11', 'pending'),
(5, 4, '2025-03-15', 'paid');

INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, line_total) VALUES
(1, 1, 1, 2, 49.00, 98.00),
(2, 1, 3, 1, 24.00, 24.00),
(3, 2, 2, 1, 129.00, 129.00),
(4, 3, 4, 3, 75.00, 225.00),
(5, 4, 1, 1, 49.00, 49.00),
(6, 5, 2, 2, 129.00, 258.00),
(7, 5, 4, 1, 75.00, 75.00);
