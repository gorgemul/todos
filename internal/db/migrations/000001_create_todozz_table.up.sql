CREATE TABLE IF NOT EXISTS todozz(
	id serial PRIMARY KEY,
	content VARCHAR(50) NOT NULL,
	created_at TIMESTAMP DEFAULT NOW()
);
