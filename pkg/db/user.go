package db

import "context"

func SavePasswordHash(ctx context.Context, hash string) error {
	_, err := DB.Exec(`
        INSERT INTO users (password_hash)
        VALUES ($1)
        ON CONFLICT (id) DO UPDATE
        SET password_hash = EXCLUDED.password_hash
    `, hash)
	return err
}

func GetUserIDByPasswordHash(ctx context.Context, hash string) (int64, error) {
	var id int64
	err := DB.QueryRowContext(ctx, `SELECT id FROM users WHERE password_hash = $1`, hash).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
