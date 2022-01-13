/* sourced template for psql pub/sub from https://webapp.io/blog/postgres-is-the-answer */

/* table definition */
CREATE TYPE yt_job_status AS ENUM ('new', 'downloading', 'downloaded', 'clipped', 'saved', 'uploaded', 'error');

CREATE TABLE yt_jobs(
	id SERIAL, 
	source_link text not null,
	start integer not null,
	duration integer not null,
	clip_link text,
	status yt_job_status, 
	status_change_time timestamp
);

/* create a job */
INSERT INTO yt_jobs(source_link, start, duration, status, status_change_time) VALUES ('JaORmA0E42A', 60, 30, 'new', NOW());

/* claim a job */
UPDATE yt_jobs SET status='downloading'
WHERE id = (
  SELECT id
  FROM yt_jobs
  WHERE status='new'
  ORDER BY id
  FOR UPDATE SKIP LOCKED
  LIMIT 1
)
RETURNING *;

/* job trigger */
CREATE OR REPLACE FUNCTION yt_job_status_notify()
	RETURNS trigger AS
$$
BEGIN
	PERFORM pg_notify('yt_job_status_channel', NEW.id::text);
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER yt_job_status
	AFTER INSERT OR UPDATE OF status
	ON ci_jobs
	FOR EACH ROW
EXECUTE PROCEDURE yt_job_status_notify();
