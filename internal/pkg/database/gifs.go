package database

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

type GifsStorage struct {
	*pgxpool.Pool
}

func NewGifsStorage(db *pgxpool.Pool) *GifsStorage {
	return &GifsStorage{db}
}

type Gif struct {
	ID  uint64
	Gif string
}

func (g *GifsStorage) CountAllGifs(ctx context.Context) (int, error) {
	row := g.QueryRow(
		ctx,
		"SELECT COUNT (*) FROM gifs",
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "cannot count all gifs")
	}

	return count, nil
}

func (g *GifsStorage) GetGifWithOffset(ctx context.Context, offset int) (Gif, error) {
	row := g.QueryRow(ctx, "SELECT id, gifid FROM gifs LIMIT 1 OFFSET $1", offset)

	gif := Gif{}
	if err := row.Scan(&gif.ID, &gif.Gif); err != nil {
		return Gif{}, err
	}

	return gif, nil
}

func (g *GifsStorage) InsertGif(ctx context.Context, gifID string) error {
	row := g.QueryRow(
		ctx,
		"SELECT COUNT (*) FROM gifs WHERE gifid = $1",
		gifID,
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return nil
	}

	if count != 0 {
		return nil
	}

	row = g.QueryRow(
		ctx,
		"INSERT INTO gifs (gifid) VALUES ($1) RETURNING id",
		gifID,
	)

	var ID uint64
	err := row.Scan(&ID)
	if err != nil {
		return errors.Wrap(err, "cannot add new gif")
	}

	return nil
}
