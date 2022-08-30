-- v15: Add guid2 column

ALTER TABLE message ADD COLUMN guid2 TEXT UNIQUE;
