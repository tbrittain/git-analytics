package query

import (
	"database/sql"
	"time"
)

// DashboardStats holds aggregate counts for the dashboard summary cards.
type DashboardStats struct {
	Commits      int `json:"commits"`
	Contributors int `json:"contributors"`
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
	FilesChanged int `json:"files_changed"`
}

// GetDashboardStats returns aggregate commit and file-change stats between
// from (inclusive) and to (exclusive).
func GetDashboardStats(db *sql.DB, from, to time.Time) (*DashboardStats, error) {
	var s DashboardStats

	err := db.QueryRow(
		`SELECT COUNT(*), COUNT(DISTINCT author_email)
		 FROM commits
		 WHERE committed_at >= ? AND committed_at < ?`,
		from, to,
	).Scan(&s.Commits, &s.Contributors)
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(
		`SELECT COALESCE(SUM(fs.additions), 0),
		        COALESCE(SUM(fs.deletions), 0),
		        COUNT(DISTINCT fs.file_path)
		 FROM file_stats fs
		 JOIN commits c ON c.hash = fs.commit_hash
		 WHERE c.committed_at >= ? AND c.committed_at < ?`,
		from, to,
	).Scan(&s.Additions, &s.Deletions, &s.FilesChanged)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// HourBucket holds the commit count for a single hour of the day (0-23).
type HourBucket struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// CommitsByHour returns per-hour commit counts between from (inclusive) and
// to (exclusive). Only hours with commits are returned (sparse).
func CommitsByHour(db *sql.DB, from, to time.Time) ([]HourBucket, error) {
	rows, err := db.Query(
		`SELECT CAST(SUBSTR(committed_at, 12, 2) AS INTEGER) AS hour,
		        COUNT(*) AS count
		 FROM commits
		 WHERE committed_at >= ? AND committed_at < ?
		 GROUP BY hour
		 ORDER BY hour`,
		from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []HourBucket
	for rows.Next() {
		var b HourBucket
		if err := rows.Scan(&b.Hour, &b.Count); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}
