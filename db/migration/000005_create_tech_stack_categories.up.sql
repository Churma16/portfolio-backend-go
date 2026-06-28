CREATE TABLE IF NOT EXISTS tech_stack_categories (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR     NOT NULL,
    slug       VARCHAR     NOT NULL UNIQUE,
    color      VARCHAR,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE tech_stacks DROP COLUMN IF EXISTS category_id;
ALTER TABLE tech_stacks ADD COLUMN IF NOT EXISTS tech_stack_category_id BIGINT REFERENCES tech_stack_categories (id) ON DELETE SET NULL;
