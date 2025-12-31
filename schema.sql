CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY, 
    type VARCHAR(255) NOT NULL, 
    payload BYTEA NOT NULL,             -- BYTEA is Postgres's version of []byte
    status VARCHAR(50) NOT NULL,
    execute_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
    );

CREATE INDEX idx_status ON tasks(status);

-- Command to run the migration 
-- docker exec -i go-dist-scheduler-postgres-1 psql -U user -d scheduler_db < schema.sql 

-- Command to list all items in database 
-- docker exec -it go-dist-scheduler-postgres-1 psql -U user -d scheduler_db -c "SELECT id, type, status, execute_at FROM tasks ORDER BY execute_at DESC;"