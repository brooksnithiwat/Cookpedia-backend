INSERT INTO categories_tag (category_tag_id, category_tag_name) VALUES
(1, 'One-dish'),
(2, 'Spicy'),
(3, 'Quick(<15 min)'),
(4, 'Vegetarian'),
(5, 'Healthy'),
(6, 'Drinks'),
(7, 'Snacks'),
(8, 'Dessert'),
(9, 'Halal'),
(10, 'Seafood'),
(11, 'Noodles'),
(12, 'Rice')
ON CONFLICT (category_tag_id) DO NOTHING;


INSERT INTO ingredients_tag (ingredient_tag_id, ingredient_tag_name)
VALUES
    (1, 'Vegetable'), 
    (2, 'Fruit'),
    (3, 'Meat'),
    (4, 'Seafood'),
    (5, 'Poultry'),
    (6, 'Dairy'),
    (7, 'Egg'),
    (8, 'Grain'),
    (9, 'Legume'),
    (10, 'Nuts & Seeds'),
    (11, 'Herbs'),
    (12, 'Spice'),
    (13, 'Oil & Fat'),
    (14, 'Sugar & Sweetener'),
    (15, 'Beverage'),      
    (16, 'Condiment'),     
    (17, 'Mushroom'),
    (18, 'Fungus & Seaweed'),
    (19, 'Baking Ingredient'),
    (20, 'Alcohol')
ON CONFLICT (ingredient_tag_name) DO NOTHING;

INSERT INTO badges (badge_id, name) VALUES
(1, 'first post'),
(2, 'got 5 star rate'),
(3, '10 recipe posted'),
(4, '25 recipe posted')
