# Distributed Task Scheduler

A distributed task scheduling system built with Go, Redis, RabbitMQ, and PostgreSQL. Ideally suited for handling delayed jobs and background processing in a scalable microservices architecture.

## Architecture

The system consists of four main components:
1.  **API Service:** A RESTful interface (Gin) that accepts task scheduling requests and stores them in Redis.
2.  **Redis:** Acts as a high-performance buffer for delayed tasks, sorted by execution time.
3.  **Scheduler:** A polling service that checks Redis for due tasks and pushes them to the RabbitMQ message broker.
4.  **Worker:** A scalable consumer service that pulls tasks from RabbitMQ and executes them. Status updates are persisted in PostgreSQL.

## Tech Stack

* **Language:** Go (Golang) 1.24
* **Database:** PostgreSQL 15
* **Message Broker:** RabbitMQ 3
* **Cache/Buffer:** Redis 7
* **Containerization:** Docker & Docker Compose

## Prerequisites

* Docker
* Docker Compose
* Git

## Installation & Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/student9876/go-dist-scheduler.git
    cd go-dist-scheduler
    ```

2.  **Start the services:**
    This command builds the images and starts all services in detached mode.
    ```bash
    docker-compose up -d --build
    ```

3.  **Initialize the Database:**
    Apply the schema to create the necessary tables.
    ```bash
    docker exec -i go-dist-scheduler-postgres-1 psql -U user -d scheduler_db < schema.sql
    ```

4.  **Verify the system is running:**
    Check the logs to ensure all services connected successfully.
    ```bash
    docker-compose logs -f
    ```

## Usage

### 1. Schedule a Single Task
Send a POST request to the API to schedule a task.

**Endpoint:** `POST http://localhost:8080/schedule`

**Example (Schedule for 10 seconds in the future):**
```bash
curl -X POST http://localhost:8080/schedule \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "payload": { "subject": "Welcome Email" },
    "execute_at": "'$(date -u -d '+10 seconds' +%Y-%m-%dT%H:%M:%SZ)'"
  }'
```

### 2. Schedule a Recurring Task
To schedule a task that repeats at a fixed interval, use the `cron` syntax.

**Endpoint:** `POST http://localhost:8080/schedule`

**Example (Every minute):**
```bash
curl -X POST http://localhost:8080/schedule \
  -H "Content-Type: application/json" \
  -d '{
    "type": "report",
    "payload": { "format": "pdf" },
    "cron": "* * * * *"
  }'
```

## 3. Scaling Workers

The system supports horizontal scaling to handle high load. Run the following command to spin up 5 concurrent workers:

```bash
docker-compose up -d --scale worker=5
```

## 4. Stress Testing (Load Generation)

You can use the following script to generate multiple tasks simultaneously and test the distributed workers. Copy and run this in your terminal:

```bash
for i in {1..10}; do
  TIMESTAMP=$(date -u -d '+10 seconds' +%Y-%m-%dT%H:%M:%SZ)
  
  curl -X POST http://localhost:8080/schedule \
    -H "Content-Type: application/json" \
    -d '{"type":"email","payload":{"subject":"Load Test '"$i"'"},"execute_at":"'"$TIMESTAMP"'"}'
    
  echo " Sent Task $i"
done
```

## Database Schema

The `tasks` table in PostgreSQL tracks the lifecycle of every job.

| Column      | Type      | Description                           |
|-------------|-----------|---------------------------------------|
| id          | UUID      | Unique Task ID                        |
| type        | VARCHAR   | Task category                         |
| status      | VARCHAR   | PENDING, RUNNING, COMPLETED, FAILED   |
| payload     | BYTEA     | Custom task data                      |
| execute_at  | TIMESTAMP | Scheduled execution time              |
| created_at  | TIMESTAMP | Task creation time                    |
