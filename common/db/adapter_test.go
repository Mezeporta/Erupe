package db

import (
	"testing"
)

func TestAdaptSQL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "placeholder rebind",
			input: "SELECT * FROM users WHERE id=$1 AND name=$2",
			want:  "SELECT * FROM users WHERE id=? AND name=?",
		},
		{
			name:  "now() replacement",
			input: "UPDATE characters SET guild_post_checked=now() WHERE id=$1",
			want:  "UPDATE characters SET guild_post_checked=CURRENT_TIMESTAMP WHERE id=?",
		},
		{
			name:  "type cast removal",
			input: "UPDATE users SET frontier_points=frontier_points::int - $1 WHERE id=$2 RETURNING frontier_points",
			want:  "UPDATE users SET frontier_points=frontier_points - ? WHERE id=? RETURNING frontier_points",
		},
		{
			name:  "text cast removal",
			input: "SELECT COALESCE(friends, ''::text) FROM characters WHERE id=$1",
			want:  "SELECT COALESCE(friends, '') FROM characters WHERE id=?",
		},
		{
			name:  "timestamptz cast removal",
			input: "SELECT COALESCE(created_at, '2000-01-01'::timestamptz) FROM guilds WHERE id=$1",
			want:  "SELECT COALESCE(created_at, '2000-01-01') FROM guilds WHERE id=?",
		},
		{
			name:  "ILIKE to LIKE",
			input: "SELECT * FROM characters WHERE name ILIKE $1",
			want:  "SELECT * FROM characters WHERE name LIKE ?",
		},
		{
			name:  "character varying cast",
			input: "DEFAULT ''::character varying",
			want:  "DEFAULT ''",
		},
		{
			name:  "no changes for standard SQL",
			input: "SELECT COUNT(*) FROM users",
			want:  "SELECT COUNT(*) FROM users",
		},
		{
			name:  "NOW uppercase",
			input: "INSERT INTO events (start_time) VALUES (NOW())",
			want:  "INSERT INTO events (start_time) VALUES (CURRENT_TIMESTAMP)",
		},
		{
			name:  "multi-digit params",
			input: "INSERT INTO t (a,b,c) VALUES ($1,$2,$10)",
			want:  "INSERT INTO t (a,b,c) VALUES (?,?,?)",
		},
		{
			name:  "public schema prefix",
			input: "INSERT INTO public.distributions_accepted VALUES ($1, $2)",
			want:  "INSERT INTO distributions_accepted VALUES (?, ?)",
		},
		{
			name:  "TRUNCATE to DELETE FROM",
			input: "TRUNCATE public.cafebonus;",
			want:  "DELETE FROM cafebonus;",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AdaptSQL(tc.input)
			if got != tc.want {
				t.Errorf("\ngot:  %s\nwant: %s", got, tc.want)
			}
		})
	}
}
