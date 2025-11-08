-- 修改 users 表：将有 DEFAULT 的字段改为 NOT NULL
ALTER TABLE users
    ALTER COLUMN gender SET DEFAULT 'unknown',
ALTER COLUMN gender SET NOT NULL,

      ALTER COLUMN status SET DEFAULT 'active',
      ALTER COLUMN status SET NOT NULL,

      ALTER COLUMN is_email_verified SET DEFAULT false,
      ALTER COLUMN is_email_verified SET NOT NULL,

      ALTER COLUMN is_phone_verified SET DEFAULT false,
      ALTER COLUMN is_phone_verified SET NOT NULL;

  -- 修改 products 表
ALTER TABLE products
    ALTER COLUMN status SET DEFAULT 'draft',
ALTER COLUMN status SET NOT NULL,

      ALTER COLUMN is_featured SET DEFAULT false,
      ALTER COLUMN is_featured SET NOT NULL,

      ALTER COLUMN low_stock_threshold SET DEFAULT 10,
      ALTER COLUMN low_stock_threshold SET NOT NULL,

      ALTER COLUMN sales_count SET DEFAULT 0,
      ALTER COLUMN sales_count SET NOT NULL,

      ALTER COLUMN view_count SET DEFAULT 0,
      ALTER COLUMN view_count SET NOT NULL;

  -- 修改 categories 表
ALTER TABLE categories
    ALTER COLUMN sort SET DEFAULT 0,
ALTER COLUMN sort SET NOT NULL,

      ALTER COLUMN level SET DEFAULT 1,
      ALTER COLUMN level SET NOT NULL,

      ALTER COLUMN is_active SET DEFAULT true,
      ALTER COLUMN is_active SET NOT NULL;

  -- 修改 carts 表
ALTER TABLE carts
    ALTER COLUMN selected SET DEFAULT true,
ALTER COLUMN selected SET NOT NULL;

  -- 修改 orders 表
ALTER TABLE orders
    ALTER COLUMN discount_amount SET DEFAULT 0,
ALTER COLUMN discount_amount SET NOT NULL,

      ALTER COLUMN shipping_fee SET DEFAULT 0,
      ALTER COLUMN shipping_fee SET NOT NULL;