ALTER TABLE tech_stacks DROP COLUMN IF EXISTS tech_stack_category_id;
ALTER TABLE tech_stacks ADD COLUMN category_id BIGINT REFERENCES categories (id) ON DELETE SET NULL;
DROP TABLE IF EXISTS tech_stack_categories;
