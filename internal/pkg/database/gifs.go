package database

import (
	"context"
	"fmt"
	"strings"

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

func (g *GifsStorage) ReplaceGifByID(ctx context.Context, gifID uint64, newGIF string) error {
	commandTag, err := g.Exec(
		ctx,
		"UPDATE gifs SET gifid = $1 where id = $2",
		gifID,
		newGIF,
	)

	if err != nil {
		return errors.Wrap(err, "cannot update gif in db")
	}

	if commandTag.RowsAffected() != 1 {
		return errors.New(
			fmt.Sprintf(
				"Rows affected %d but excepted 1",
				commandTag.RowsAffected(),
			),
		)
	}

	return nil
}

func (g *GifsStorage) GetRandomGifs(ctx context.Context, randomIDs []string) ([]string, error) {
	gifs := make([]string, len(randomIDs))

	rows, err := g.Query(
		ctx,
		fmt.Sprintf(
			"SELECT gifid FROM (SELECT gifid, ROW_NUMBER() OVER(ORDER BY id) AS index FROM gifs) as t where index  IN (%v)",
			strings.Join(randomIDs, ", "),
		),
	)

	if err != nil {
		return nil, errors.Wrap(err, "cannot select many random gifs")
	}

	scanIndex := 0
	for rows.Next() {
		err := rows.Scan(&gifs[scanIndex])
		scanIndex++

		if err != nil {
			return nil, errors.Wrap(err, "cannot scan to to gifs array")
		}
	}

	return gifs, nil
}

func (g *GifsStorage) InsertGif(ctx context.Context, gifID string) error {
	eqID := "%" + gifID[len(gifID)-21:]
	row := g.QueryRow(
		ctx,
		"SELECT COUNT (*) FROM gifs WHERE gifid LIKE $1",
		eqID,
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

func (g *GifsStorage) DeleteGifByFileID(ctx context.Context, fileID string) error {
	eqID := "%" + fileID[len(fileID)-21:]

	commandTag, err := g.Exec(
		ctx,
		"DELETE FROM gifs WHERE gifid LIKE $1",
		eqID,
	)

	if err != nil {
		return errors.Wrap(err, "cannot delete gif")
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New("no row found to delete")
	}

	return nil
}
