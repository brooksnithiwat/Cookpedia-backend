-- SET TIMEZONE = 'Asia/Bangkok';


-- CREATE TABLE IF NOT EXISTS categories_tag (
--     category_tag_id SERIAL PRIMARY KEY,
--     category_tag_name VARCHAR(25) UNIQUE NOT NULL
-- );


-- CREATE TABLE IF NOT EXISTS ingredients_tag (
--     ingredient_tag_id SERIAL PRIMARY KEY,
--     ingredient_tag_name VARCHAR(25) UNIQUE NOT NULL
-- );

-- CREATE TABLE IF NOT EXISTS users (
--     user_id SERIAL PRIMARY KEY,
--     username VARCHAR(30) NOT NULL,
--     email VARCHAR(30) UNIQUE NOT NULL,
--     password VARCHAR(255),
--     google_id VARCHAR(255) UNIQUE,
--     provider VARCHAR(50) NOT NULL DEFAULT 'local',
--     role VARCHAR(50) NOT NULL DEFAULT 'user',
--     firstname VARCHAR(30),
--     lastname VARCHAR(30),
--     phone VARCHAR(30),
--     aboutme TEXT,
--     image_url TEXT,
--     CONSTRAINT check_provider_values CHECK (provider IN ('local', 'google')),
--     CONSTRAINT check_local_user_password CHECK (
--         (provider = 'local' AND password IS NOT NULL) OR 
--         (provider != 'local')
--     ),
--     CONSTRAINT check_google_user_google_id CHECK (
--         (provider = 'google' AND google_id IS NOT NULL) OR 
--         (provider != 'google')
--     )
-- );

-- CREATE TABLE IF NOT EXISTS posts (
--     post_id SERIAL PRIMARY KEY,
--     user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
--     menu_name VARCHAR(255) NOT NULL,
--     story TEXT,
--     image_url TEXT,
--     created_at TIMESTAMP NOT NULL DEFAULT NOW()  
-- );


-- CREATE TABLE post_categories (
--     post_id INT NOT NULL REFERENCES posts(post_id) ON DELETE CASCADE,
--     category_tag_id INT NOT NULL REFERENCES categories_tag(category_tag_id) ON DELETE CASCADE,
--     PRIMARY KEY (post_id, category_tag_id)
-- );

-- CREATE TABLE post_ingredients (
--     post_id INT NOT NULL REFERENCES posts(post_id) ON DELETE CASCADE,
--     ingredient_tag_id INT NOT NULL REFERENCES ingredients_tag(ingredient_tag_id) ON DELETE CASCADE,
--     PRIMARY KEY (post_id, ingredient_tag_id)
-- );


-- CREATE TABLE IF NOT EXISTS ingredients_detail (
--     post_id INT NOT NULL REFERENCES posts(post_id) ON DELETE CASCADE,
--     detail TEXT NOT NULL
-- );


-- CREATE TABLE IF NOT EXISTS instructions (
--     post_id INT NOT NULL REFERENCES posts(post_id) ON DELETE CASCADE,
--     step_number INT NOT NULL,
--     detail TEXT NOT NULL,
--     CONSTRAINT uq_post_step UNIQUE (post_id, step_number)
-- );


-- CREATE TABLE IF NOT EXISTS favorite (
--     user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
--     post_id INT NOT NULL REFERENCES posts(post_id) ON DELETE CASCADE,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     PRIMARY KEY (user_id, post_id)
-- );


-- CREATE TABLE IF NOT EXISTS rating (
--     rating_id SERIAL PRIMARY KEY,
--     user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
--     post_id INT NOT NULL REFERENCES posts(post_id) ON DELETE CASCADE,
--     rate INT NOT NULL CHECK (rate >= 1 AND rate <= 5),
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     CONSTRAINT uq_user_post UNIQUE (user_id, post_id)
-- );

-- CREATE TABLE IF NOT EXISTS comment (
--     comment_id SERIAL PRIMARY KEY,
--     user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
--     post_id INT NOT NULL REFERENCES posts(post_id) ON DELETE CASCADE,
--     content TEXT NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );



-- CREATE INDEX idx_post_categories_post_id ON post_categories(post_id);
-- CREATE INDEX idx_post_categories_category_tag_id ON post_categories(category_tag_id);
-- CREATE INDEX idx_post_ingredients_post_id ON post_ingredients(post_id);
-- CREATE INDEX idx_post_ingredients_ingredient_tag_id ON post_ingredients(ingredient_tag_id);
-- CREATE INDEX idx_ingredients_detail_post_id ON ingredients_detail(post_id);
-- CREATE INDEX idx_instructions_post_id ON instructions(post_id);


CREATE TABLE IF NOT EXISTS followers (
    follower_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    following_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (follower_id, following_id)
);