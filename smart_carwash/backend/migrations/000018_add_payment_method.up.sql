-- Add payment_method column to payments table
ALTER TABLE payments ADD COLUMN payment_method VARCHAR(50) DEFAULT 'tinkoff'; 