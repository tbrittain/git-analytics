package query

import (
	"database/sql"
	"math"
	"sort"
	"strings"
	"time"
)

// FileHotspot represents aggregated churn for a single file path.
type FileHotspot struct {
	Path         string `json:"path"`
	LinesChanged int    `json:"lines_changed"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	Commits      int    `json:"commits"`
}

// FileHotspots returns per-file churn (additions + deletions) and commit counts
// for commits between from (inclusive) and to (exclusive), ordered by
// lines_changed descending. Files matching any of the excludeGlobs patterns
// are omitted from results entirely.
func FileHotspots(db *sql.DB, from, to time.Time, excludeGlobs []string) ([]FileHotspot, error) {
	excludeSQL, excludeArgs := buildExcludeClauses("fs.file_path", excludeGlobs)

	q := `SELECT fs.file_path,
	        SUM(fs.additions + fs.deletions) AS lines_changed,
	        SUM(fs.additions) AS additions,
	        SUM(fs.deletions) AS deletions,
	        COUNT(DISTINCT fs.commit_hash) AS commits
	 FROM file_stats fs
	 JOIN commits c ON c.hash = fs.commit_hash
	 WHERE c.committed_at >= ? AND c.committed_at < ?` + excludeSQL + `
	 GROUP BY fs.file_path
	 ORDER BY lines_changed DESC`

	args := make([]any, 0, len(excludeArgs)+2)
	args = append(args, from, to)
	args = append(args, excludeArgs...)

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []FileHotspot
	for rows.Next() {
		var h FileHotspot
		if err := rows.Scan(&h.Path, &h.LinesChanged, &h.Additions, &h.Deletions, &h.Commits); err != nil {
			return nil, err
		}
		result = append(result, h)
	}
	return result, rows.Err()
}

// TemporalHotspot extends FileHotspot with recency-weighted scoring.
type TemporalHotspot struct {
	Path         string  `json:"path"`
	LinesChanged int     `json:"lines_changed"`
	Additions    int     `json:"additions"`
	Deletions    int     `json:"deletions"`
	Commits      int     `json:"commits"`
	LastChanged  string  `json:"last_changed"`
	DaysSince    int     `json:"days_since"`
	Score        float64 `json:"score"`
}

// TemporalHotspots returns per-file churn weighted by recency using exponential
// decay: score = lines_changed * e^(-λ * daysSince) where λ = ln(2)/halfLifeDays.
// Results are ordered by score descending. The reference time for recency is `to`.
func TemporalHotspots(db *sql.DB, from, to time.Time, halfLifeDays float64, excludeGlobs []string) ([]TemporalHotspot, error) {
	excludeSQL, excludeArgs := buildExcludeClauses("fs.file_path", excludeGlobs)

	q := `SELECT fs.file_path,
	        SUM(fs.additions + fs.deletions) AS lines_changed,
	        SUM(fs.additions) AS additions,
	        SUM(fs.deletions) AS deletions,
	        COUNT(DISTINCT fs.commit_hash) AS commits,
	        MAX(c.committed_at) AS last_committed_at
	 FROM file_stats fs
	 JOIN commits c ON c.hash = fs.commit_hash
	 WHERE c.committed_at >= ? AND c.committed_at < ?` + excludeSQL + `
	 GROUP BY fs.file_path`

	args := make([]any, 0, len(excludeArgs)+2)
	args = append(args, from, to)
	args = append(args, excludeArgs...)

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lambda := math.Ln2 / halfLifeDays

	var result []TemporalHotspot
	for rows.Next() {
		var h TemporalHotspot
		var lastCommittedAt string
		if err := rows.Scan(&h.Path, &h.LinesChanged, &h.Additions, &h.Deletions, &h.Commits, &lastCommittedAt); err != nil {
			return nil, err
		}

		lastTime, err := time.Parse(time.RFC3339, lastCommittedAt)
		if err != nil {
			// modernc.org/sqlite serializes time.Time via Go's String() method:
			// "2006-01-02 15:04:05 +0000 UTC" or "2006-01-02 15:04:05 +0900 +0900".
			// Strip the trailing tz name so we can parse the numeric offset alone.
			trimmed := lastCommittedAt
			if idx := strings.LastIndex(trimmed, " "); idx > 0 {
				trimmed = trimmed[:idx]
			}
			lastTime, err = time.Parse("2006-01-02 15:04:05 -0700", trimmed)
			if err != nil {
				return nil, err
			}
		}

		daysSince := to.Sub(lastTime).Hours() / 24
		if daysSince < 0 {
			daysSince = 0
		}
		h.DaysSince = int(daysSince)
		h.LastChanged = lastTime.Format("2006-01-02")
		h.Score = float64(h.LinesChanged) * math.Exp(-lambda*daysSince)

		result = append(result, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	return result, nil
}
