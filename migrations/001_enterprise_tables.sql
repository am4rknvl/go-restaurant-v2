-- Enterprise tables for restaurant platform

-- User roles table
CREATE TABLE IF NOT EXISTS user_roles (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    role TEXT NOT NULL,
    restaurant_id TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_user_roles_account_id ON user_roles(account_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_restaurant_id ON user_roles(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_deleted_at ON user_roles(deleted_at);

-- Inventory items table
CREATE TABLE IF NOT EXISTS inventory_items (
    id TEXT PRIMARY KEY,
    restaurant_id TEXT NOT NULL,
    sku TEXT NOT NULL,
    name TEXT NOT NULL,
    qty DECIMAL NOT NULL DEFAULT 0,
    unit TEXT,
    reorder_level DECIMAL DEFAULT 0,
    cost DECIMAL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_inventory_items_restaurant_id ON inventory_items(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_inventory_items_sku ON inventory_items(sku);
CREATE INDEX IF NOT EXISTS idx_inventory_items_deleted_at ON inventory_items(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_items_restaurant_sku ON inventory_items(restaurant_id, sku) WHERE deleted_at IS NULL;

-- Inventory adjustments table
CREATE TABLE IF NOT EXISTS inventory_adjustments (
    id TEXT PRIMARY KEY,
    item_id TEXT NOT NULL REFERENCES inventory_items(id),
    delta DECIMAL NOT NULL,
    reason TEXT,
    user_id TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_inventory_adjustments_item_id ON inventory_adjustments(item_id);
CREATE INDEX IF NOT EXISTS idx_inventory_adjustments_created_at ON inventory_adjustments(created_at);

-- Staff assignments table
CREATE TABLE IF NOT EXISTS staff_assignments (
    id TEXT PRIMARY KEY,
    restaurant_id TEXT NOT NULL,
    staff_id TEXT NOT NULL,
    table_id TEXT,
    order_id TEXT,
    assign_type TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_staff_assignments_restaurant_id ON staff_assignments(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_staff_assignments_staff_id ON staff_assignments(staff_id);
CREATE INDEX IF NOT EXISTS idx_staff_assignments_table_id ON staff_assignments(table_id);
CREATE INDEX IF NOT EXISTS idx_staff_assignments_order_id ON staff_assignments(order_id);

-- Order audit table
CREATE TABLE IF NOT EXISTS order_audits (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL,
    action TEXT NOT NULL,
    user_id TEXT,
    details TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_order_audits_order_id ON order_audits(order_id);
CREATE INDEX IF NOT EXISTS idx_order_audits_created_at ON order_audits(created_at);

-- Discounts table
CREATE TABLE IF NOT EXISTS discounts (
    id TEXT PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    type TEXT NOT NULL,
    value DECIMAL NOT NULL,
    restaurant_id TEXT,
    valid_from TIMESTAMPTZ DEFAULT NOW(),
    valid_to TIMESTAMPTZ,
    usage_limit INTEGER DEFAULT 0,
    per_user_limit INTEGER DEFAULT 1,
    used_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_discounts_code ON discounts(code);
CREATE INDEX IF NOT EXISTS idx_discounts_restaurant_id ON discounts(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_discounts_valid_from ON discounts(valid_from);
CREATE INDEX IF NOT EXISTS idx_discounts_valid_to ON discounts(valid_to);
CREATE INDEX IF NOT EXISTS idx_discounts_deleted_at ON discounts(deleted_at);

-- Discount usage table
CREATE TABLE IF NOT EXISTS discount_usages (
    id TEXT PRIMARY KEY,
    discount_id TEXT NOT NULL REFERENCES discounts(id),
    account_id TEXT NOT NULL,
    order_id TEXT,
    amount DECIMAL NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_discount_usages_discount_id ON discount_usages(discount_id);
CREATE INDEX IF NOT EXISTS idx_discount_usages_account_id ON discount_usages(account_id);

-- Loyalty accounts table
CREATE TABLE IF NOT EXISTS loyalty_accounts (
    id TEXT PRIMARY KEY,
    account_id TEXT UNIQUE NOT NULL,
    points INTEGER DEFAULT 0,
    tier TEXT DEFAULT 'bronze',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_loyalty_accounts_account_id ON loyalty_accounts(account_id);

-- Loyalty transactions table
CREATE TABLE IF NOT EXISTS loyalty_transactions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    points INTEGER NOT NULL,
    type TEXT NOT NULL,
    order_id TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_loyalty_transactions_account_id ON loyalty_transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_loyalty_transactions_created_at ON loyalty_transactions(created_at);

-- Restaurants table
CREATE TABLE IF NOT EXISTS restaurants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    timezone TEXT DEFAULT 'UTC',
    currency TEXT DEFAULT 'USD',
    tax_rate DECIMAL DEFAULT 0,
    address TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_restaurants_deleted_at ON restaurants(deleted_at);

-- Table states table
CREATE TABLE IF NOT EXISTS table_states (
    id TEXT PRIMARY KEY,
    table_id TEXT NOT NULL,
    state TEXT NOT NULL,
    assigned_to TEXT,
    restaurant_id TEXT NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_table_states_table_id ON table_states(table_id);
CREATE INDEX IF NOT EXISTS idx_table_states_restaurant_id ON table_states(restaurant_id);

-- Waitlist entries table
CREATE TABLE IF NOT EXISTS waitlist_entries (
    id TEXT PRIMARY KEY,
    restaurant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    phone TEXT,
    party_size INTEGER NOT NULL,
    status TEXT DEFAULT 'waiting',
    position INTEGER DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_waitlist_entries_restaurant_id ON waitlist_entries(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_waitlist_entries_status ON waitlist_entries(status);
CREATE INDEX IF NOT EXISTS idx_waitlist_entries_position ON waitlist_entries(position);
CREATE INDEX IF NOT EXISTS idx_waitlist_entries_deleted_at ON waitlist_entries(deleted_at);

-- Payment tips table
CREATE TABLE IF NOT EXISTS payment_tips (
    id TEXT PRIMARY KEY,
    payment_id TEXT NOT NULL,
    amount DECIMAL NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_payment_tips_payment_id ON payment_tips(payment_id);

-- Add restaurant_id to existing tables if not exists
ALTER TABLE IF EXISTS accounts ADD COLUMN IF NOT EXISTS restaurant_id TEXT;
ALTER TABLE IF EXISTS orders ADD COLUMN IF NOT EXISTS restaurant_id TEXT;
ALTER TABLE IF EXISTS menu_items ADD COLUMN IF NOT EXISTS restaurant_id TEXT;
ALTER TABLE IF EXISTS tables ADD COLUMN IF NOT EXISTS restaurant_id TEXT;
ALTER TABLE IF EXISTS reservations ADD COLUMN IF NOT EXISTS restaurant_id TEXT;

-- Add indexes for restaurant_id on existing tables
CREATE INDEX IF NOT EXISTS idx_accounts_restaurant_id ON accounts(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_orders_restaurant_id ON orders(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_menu_items_restaurant_id ON menu_items(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_tables_restaurant_id ON tables(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_reservations_restaurant_id ON reservations(restaurant_id);

-- Performance indexes for reporting
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_order_items_menu_item_id ON order_items(menu_item_id);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);