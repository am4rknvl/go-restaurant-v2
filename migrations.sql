-- Basic migration file for restaurant system (partial)
-- Run these statements in your Postgres DB (psql -f migrations.sql)

CREATE TABLE IF NOT EXISTS categories (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS menu_items (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  price NUMERIC NOT NULL,
  category TEXT,
  available BOOLEAN DEFAULT TRUE,
  image_url TEXT,
  special_notes TEXT
);

CREATE TABLE IF NOT EXISTS tables (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  qr_code TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,
  table_id TEXT NOT NULL REFERENCES tables(id),
  customer TEXT,
  status TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  closed_at TIMESTAMP WITH TIME ZONE
);

-- Note: orders, order_items, payments tables likely exist in your project; adjust as needed.

-- Ensure accounts have role column
ALTER TABLE IF EXISTS accounts ADD COLUMN IF NOT EXISTS role TEXT DEFAULT 'customer';

-- Ensure orders have session_id for session linking
ALTER TABLE IF EXISTS orders ADD COLUMN IF NOT EXISTS session_id TEXT;

-- Enhance menu_items with multilingual fields, images and special_notes
ALTER TABLE IF EXISTS menu_items ADD COLUMN IF NOT EXISTS image_url TEXT;
ALTER TABLE IF EXISTS menu_items ADD COLUMN IF NOT EXISTS special_notes TEXT;
ALTER TABLE IF EXISTS menu_items ADD COLUMN IF NOT EXISTS name_am TEXT;
ALTER TABLE IF EXISTS menu_items ADD COLUMN IF NOT EXISTS description_am TEXT;

-- Add special_instructions to order_items
ALTER TABLE IF EXISTS order_items ADD COLUMN IF NOT EXISTS special_instructions TEXT;

-- Favorites table: customers can favorite items
CREATE TABLE IF NOT EXISTS favorites (
  id TEXT PRIMARY KEY,
  account_id TEXT NOT NULL,
  menu_item_id TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Reviews table: ratings and comments per menu item
CREATE TABLE IF NOT EXISTS reviews (
  id TEXT PRIMARY KEY,
  account_id TEXT NOT NULL,
  menu_item_id TEXT NOT NULL,
  rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
  comment TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Reservations for tables
CREATE TABLE IF NOT EXISTS reservations (
  id TEXT PRIMARY KEY,
  account_id TEXT NOT NULL,
  table_id TEXT NOT NULL REFERENCES tables(id),
  party_size INT DEFAULT 1,
  reserved_for TIMESTAMP WITH TIME ZONE NOT NULL,
  status TEXT DEFAULT 'booked', -- booked, cancelled, completed
  notes TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Notifications / subscriptions for push or SMS
CREATE TABLE IF NOT EXISTS subscriptions (
  id TEXT PRIMARY KEY,
  account_id TEXT NOT NULL,
  kind TEXT NOT NULL, -- push | sms
  endpoint TEXT NOT NULL,
  metadata JSONB,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Extend orders with ETA and prepared timestamp
ALTER TABLE IF EXISTS orders ADD COLUMN IF NOT EXISTS estimated_ready_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE IF EXISTS orders ADD COLUMN IF NOT EXISTS prepared_at TIMESTAMP WITH TIME ZONE;

-- Payments: support partial payments and refunds tracking
ALTER TABLE IF EXISTS payments ADD COLUMN IF NOT EXISTS refunded_amount NUMERIC DEFAULT 0;
ALTER TABLE IF EXISTS payments ADD COLUMN IF NOT EXISTS is_partial BOOLEAN DEFAULT FALSE;
CREATE TABLE IF NOT EXISTS refunds (
  id TEXT PRIMARY KEY,
  payment_id TEXT NOT NULL REFERENCES payments(id),
  amount NUMERIC NOT NULL,
  reason TEXT,
  status TEXT DEFAULT 'pending', -- pending, completed, rejected
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Branches for multi-restaurant support
CREATE TABLE IF NOT EXISTS branches (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  timezone TEXT,
  currency TEXT,
  tax_rate NUMERIC,
  address TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Add branch_id to key entities if not exists
ALTER TABLE IF EXISTS tables ADD COLUMN IF NOT EXISTS branch_id TEXT REFERENCES branches(id);
ALTER TABLE IF EXISTS menu_items ADD COLUMN IF NOT EXISTS branch_id TEXT REFERENCES branches(id);
ALTER TABLE IF EXISTS reservations ADD COLUMN IF NOT EXISTS branch_id TEXT REFERENCES branches(id);
ALTER TABLE IF EXISTS orders ADD COLUMN IF NOT EXISTS branch_id TEXT REFERENCES branches(id);

-- Table state management
ALTER TABLE IF EXISTS tables ADD COLUMN IF NOT EXISTS state TEXT DEFAULT 'available';

-- Waitlist per branch
CREATE TABLE IF NOT EXISTS waitlist (
  id TEXT PRIMARY KEY,
  branch_id TEXT NOT NULL REFERENCES branches(id),
  name TEXT NOT NULL,
  phone TEXT,
  party_size INT NOT NULL DEFAULT 1,
  status TEXT NOT NULL DEFAULT 'waiting', -- waiting | notified | seated | expired
  position INT,
  notes TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Inventory per branch
CREATE TABLE IF NOT EXISTS inventory_items (
  id TEXT PRIMARY KEY,
  branch_id TEXT NOT NULL REFERENCES branches(id),
  sku TEXT NOT NULL,
  name TEXT NOT NULL,
  qty NUMERIC NOT NULL DEFAULT 0,
  unit TEXT,
  reorder_level NUMERIC DEFAULT 0,
  cost NUMERIC,
  UNIQUE (branch_id, sku)
);

CREATE TABLE IF NOT EXISTS inventory_adjustments (
  id TEXT PRIMARY KEY,
  item_id TEXT NOT NULL REFERENCES inventory_items(id),
  delta NUMERIC NOT NULL,
  reason TEXT,
  user_id TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Recipes link menu items to ingredients
CREATE TABLE IF NOT EXISTS recipes (
  id TEXT PRIMARY KEY,
  branch_id TEXT NOT NULL REFERENCES branches(id),
  menu_item_id TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS recipe_components (
  id TEXT PRIMARY KEY,
  recipe_id TEXT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
  sku TEXT NOT NULL,
  qty NUMERIC NOT NULL,
  unit TEXT
);