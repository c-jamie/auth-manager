package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)
// Team represents the domain for our team entity
type Team struct {
	ID           int64          `json:"id"`
	Name         string         `json:"name"`
	CreatedAt    time.Time      `json:"created_at"`
	Version      int            `json:"version"`
	NumMembers   int            `json:"num_members"`
	Meta         meta           `json:"meta"`
	UserAccounts []*UserAccount `json:"user_accounts"`
}

type meta struct {
	GitURL    string `json:"git_url"`
	ServerURL string `json:"server_url"`
	Version   int64  `json:"version"`
	ID        int64  `json:"id"`
}
// TeamModel wraps the connection pool
type TeamModel struct {
	DB *sql.DB
}

// Add adds a Team into the database
func (m TeamModel) Add(team *Team) error {
	query := `
		insert into team(name, created_at)
		values ($1 , now())
		returning id, created_at, version
	`
	args := []interface{}{team.Name}
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancle()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&team.ID, &team.CreatedAt, &team.Version)
	if err != nil {
		return err
	}

	query = `
		insert into team_meta(git_url, server_url, created_at)
		values ($1, $2, now())
		returning id
	`
	args = []interface{}{team.Meta.GitURL, team.Meta.ServerURL}
	ctx, cancle = context.WithTimeout(context.Background(), 3*time.Second)

	defer cancle()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&team.Meta.ID)
	if err != nil {
		return err
	}

	return nil
}

// Get returns the Team for a given Team ID
func (m TeamModel) Get(id int64) (*Team, error) {
	query := `
		select 		t.id, t.created_at, t.name, t.version, tm.id, tm.git_url, tm.server_url
		from 		team as t
		inner join	team_meta as tm
		on			t.team_meta_id = tm.id
		where		id = $1
	`

	var team Team

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&team.ID,
		&team.CreatedAt,
		&team.Name,
		&team.Version,
		&team.Meta.ID,
		&team.Meta.GitURL,
		&team.Meta.ServerURL,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, fmt.Errorf("no record found: %w", err)
		default:
			return nil, err
		}
	}
	return &team, nil
}

// Update updates the Team 
func (m TeamModel) Update(team *Team) error {

	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	tx, err := m.DB.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("unable to update %w", err)
	}

	query1 := `
		update 		team as t
		set 		t.name = $1
					, t.version = version + 1

		where 		t.id = $2 and t.version = $3
		returning 	version
	`
	args1 := []interface{}{
		team.Name,
		team.ID,
		team.Version,
	}

	err = tx.QueryRowContext(ctx, query1, args1...).Scan(&team.Version)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to update %w", err)
	}

	query2 := `
		update 		team_meta as tm

		set 		tm.git_url = $1
					, tm.server_url =$2

		from 		team as t
		on 			tm.id = t.team_meta_id
		where 		t.id = $3 
		and 		t.version = $3

		returning 	t.version
	`
	args2 := []interface{}{
		team.Meta.GitURL,
		team.Meta.ServerURL,
		team.ID,
		team.Version,
	}

	err = tx.QueryRowContext(ctx, query2, args2...).Scan(team.Version)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to update %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("unable to update %w", err)
	}

	return nil
}
