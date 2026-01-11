-- 1. USERS
CREATE TABLE users (
                       id BIGSERIAL PRIMARY KEY,
                       email VARCHAR NOT NULL UNIQUE,
                       password VARCHAR NOT NULL,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 2. MASTER DATA (Category, TechStack, Tag)
CREATE TABLE categories (
                            id BIGSERIAL PRIMARY KEY,
                            name VARCHAR NOT NULL,
                            slug VARCHAR NOT NULL UNIQUE,
                            color VARCHAR,
                            created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

create table messages (
                            id BIGSERIAL PRIMARY KEY,
                            name VARCHAR NOT NULL,
                            email VARCHAR NOT NULL,
                            content TEXT NOT NULL,
                            is_read BOOLEAN DEFAULT false,
                            created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE tech_stacks (
                             id BIGSERIAL PRIMARY KEY,
                             name VARCHAR NOT NULL,
                             slug VARCHAR NOT NULL UNIQUE,
                             icon VARCHAR,
                             created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- (BARU) Tabel Tags dengan relasi ke Category
CREATE TABLE tags (
                      id BIGSERIAL PRIMARY KEY,
                      name VARCHAR NOT NULL,
                      slug VARCHAR NOT NULL UNIQUE,
                      color VARCHAR,
                      category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
                      created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 3. PROFILES
CREATE TABLE profiles (
                          id BIGSERIAL PRIMARY KEY,
                          user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          name VARCHAR NOT NULL,
                          headline VARCHAR,
                          role VARCHAR,
                          bio_short TEXT,
                          bio_long TEXT,
                          location VARCHAR,
                          is_hireable BOOLEAN DEFAULT false,
                          avatar VARCHAR,
                          cv_files VARCHAR,
                          hero_image_codes VARCHAR,
                          socials JSONB,
                          created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                          updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 4. WORK EXPERIENCES
CREATE TABLE work_experiences (
                                  id BIGSERIAL PRIMARY KEY,
                                  company VARCHAR NOT NULL,
                                  position VARCHAR NOT NULL,
                                  location VARCHAR,
                                  start_date VARCHAR,
                                  end_date VARCHAR,
                                  is_current BOOLEAN DEFAULT false,
                                  description TEXT,
                                  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 5. PROJECTS
CREATE TABLE projects (
                          id BIGSERIAL PRIMARY KEY,
                          title VARCHAR NOT NULL,
                          slug VARCHAR NOT NULL UNIQUE,
                          thumbnail VARCHAR,
                          content TEXT,
                          demo_url VARCHAR,
                          repo_url VARCHAR,
                          is_featured BOOLEAN DEFAULT false,
                          published_at TIMESTAMPTZ,
                          category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
                          created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                          updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 6. PIVOT TABLES (Relasi Many-to-Many)

-- Project <-> TechStack
CREATE TABLE project_tech_stacks (
                                     project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
                                     tech_stack_id BIGINT NOT NULL REFERENCES tech_stacks(id) ON DELETE CASCADE,
                                     PRIMARY KEY (project_id, tech_stack_id)
);

-- (BARU) Project <-> Tag
CREATE TABLE project_tags (
                              project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
                              tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
                              PRIMARY KEY (project_id, tag_id)
);

-- WorkExperience <-> TechStack
CREATE TABLE work_experience_tech_stacks (
                                             work_experience_id BIGINT NOT NULL REFERENCES work_experiences(id) ON DELETE CASCADE,
                                             tech_stack_id BIGINT NOT NULL REFERENCES tech_stacks(id) ON DELETE CASCADE,
                                             PRIMARY KEY (work_experience_id, tech_stack_id)
);

-- (BARU) WorkExperience <-> Tag
CREATE TABLE work_experience_tags (
                                      work_experience_id BIGINT NOT NULL REFERENCES work_experiences(id) ON DELETE CASCADE,
                                      tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
                                      PRIMARY KEY (work_experience_id, tag_id)
);