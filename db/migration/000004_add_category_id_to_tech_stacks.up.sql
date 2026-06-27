ALTER TABLE tech_stacks ADD COLUMN category_id BIGINT REFERENCES categories (id) ON DELETE SET NULL;
