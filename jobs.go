package main

import "fmt"

// Job represents the job/database schema
type Job struct {
	Id               int    `json:"id"`
	SourceLink       string `json:"source_link"`
	Start            int64  `json:"start"`
	Duration         int64  `json:"duration"`
	ClipLink         string `json:"clip_link"`
	Status           string `json:"status"`
	StatusChangeTime string `json:"status_change_time"`
}

// Jobs retrives all from the database
func (c *Clips) Jobs() ([]Job, error) {
	var results []Job

	rows, err := c.database.Query(`SELECT id, source_link, start, duration, status, status_change_time FROM yt_jobs`)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var result Job

		err = rows.Scan(&result.Id, &result.SourceLink, &result.Start, &result.Duration, &result.Status, &result.StatusChangeTime)
		if err != nil {
			fmt.Println(err)
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

// InsertJob adds a new job to the database
func (c *Clips) InsertJob(link string, start, duration int64) (Job, error) {
	var result Job

	stmt := `
		INSERT INTO yt_jobs
			(source_link, start, duration, status, status_change_time)
		VALUES
			($1, $2, $3, 'new', NOW())
		RETURNING id`
	args := []interface{}{
		link,
		start,
		duration,
	}
	row := c.database.QueryRow(stmt, args...)

	err := row.Scan(&result.Id)
	if err != nil {
		return result, err
	}

	return result, nil
}

// FindJob inserts a new entry in the database
func (c *Clips) FindJob(link string, start, duration int64) (bool, error) {
	var result bool

	stmt := `select exists(SELECT 1 FROM yt_jobs WHERE source_link = $1 AND start = $2 AND duration = $3)`
	args := []interface{}{
		link,
		start,
		duration,
	}
	row := c.database.QueryRow(stmt, args...)

	err := row.Scan(&result)
	if err != nil {
		return result, err
	}

	return result, nil
}
