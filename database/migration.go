package database

import (
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// RunMigrations creates necessary database tables if they don't exist
func RunMigrations(db *sql.DB) error {
	// Create users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'cashier',
		is_active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.Exec(createUsersTable)
	if err != nil {
		return err
	}
	log.Println("Users table ready")

	// Seed default owner account if no users exist
	var userCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if userCount == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		_, err = db.Exec(
			"INSERT INTO users (name, email, password, role) VALUES ($1, $2, $3, $4)",
			"Admin", "admin@retail.com", string(hash), "owner",
		)
		if err != nil {
			log.Println("Warning: failed to seed admin user:", err)
		} else {
			log.Println("Default admin user seeded (admin@retail.com / password123)")
		}
	}

	// Create categories table
	createCategoriesTable := `
	CREATE TABLE IF NOT EXISTS categories (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(createCategoriesTable)
	if err != nil {
		return err
	}
	log.Println("Categories table ready")

	// Create products table with foreign key to categories
	createProductsTable := `
	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		price INTEGER NOT NULL DEFAULT 0,
		stock INTEGER NOT NULL DEFAULT 0,
		sku VARCHAR(100) DEFAULT '',
		image_url TEXT DEFAULT '',
		unit VARCHAR(50) DEFAULT 'pcs',
		is_active BOOLEAN DEFAULT true,
		category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(createProductsTable)
	if err != nil {
		return err
	}
	log.Println("Products table ready")

	// Add new columns if they don't exist (for existing databases)
	alterProducts := []string{
		"ALTER TABLE products ADD COLUMN IF NOT EXISTS sku VARCHAR(100) DEFAULT ''",
		"ALTER TABLE products ADD COLUMN IF NOT EXISTS image_url TEXT DEFAULT ''",
		"ALTER TABLE products ADD COLUMN IF NOT EXISTS unit VARCHAR(50) DEFAULT 'pcs'",
		"ALTER TABLE products ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true",
	}
	for _, q := range alterProducts {
		_, _ = db.Exec(q)
	}

	// Create index on category_id for better JOIN performance
	createIndexQuery := `
	CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
	`

	_, err = db.Exec(createIndexQuery)
	if err != nil {
		return err
	}
	log.Println("Database indexes ready")

	// Create transactions table
	createTransactionsTable := `
	CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		total_amount INT NOT NULL,
		payment_method VARCHAR(50) DEFAULT 'cash',
		discount INT DEFAULT 0,
		notes TEXT DEFAULT '',
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(createTransactionsTable)
	if err != nil {
		return err
	}
	log.Println("Transactions table ready")

	// Add new columns to transactions if they don't exist
	alterTransactions := []string{
		"ALTER TABLE transactions ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50) DEFAULT 'cash'",
		"ALTER TABLE transactions ADD COLUMN IF NOT EXISTS discount INT DEFAULT 0",
		"ALTER TABLE transactions ADD COLUMN IF NOT EXISTS notes TEXT DEFAULT ''",
		"ALTER TABLE transactions ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active'",
	}
	for _, q := range alterTransactions {
		_, _ = db.Exec(q)
	}

	// Create transaction_details table
	createTransactionDetailsTable := `
	CREATE TABLE IF NOT EXISTS transaction_details (
		id SERIAL PRIMARY KEY,
		transaction_id INT REFERENCES transactions(id) ON DELETE CASCADE,
		product_id INT REFERENCES products(id),
		quantity INT NOT NULL,
		unit_price INT NOT NULL DEFAULT 0,
		subtotal INT NOT NULL
	);
	`

	_, err = db.Exec(createTransactionDetailsTable)
	if err != nil {
		return err
	}
	log.Println("Transaction details table ready")

	// Add unit_price column if it doesn't exist
	_, _ = db.Exec("ALTER TABLE transaction_details ADD COLUMN IF NOT EXISTS unit_price INT DEFAULT 0")

	return nil
}
