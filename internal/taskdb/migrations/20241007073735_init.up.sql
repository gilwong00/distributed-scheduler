CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tasks (
	-- id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	command TEXT NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
	scheduled_at TIMESTAMP NOT NULL,
	picked_at TIMESTAMP WITH TIME ZONE,
	started_at TIMESTAMP WITH TIME ZONE, -- when the worker started executing the task.
	completed_at TIMESTAMP WITH TIME ZONE, -- when the task was completed (success case)
	failed_at TIMESTAMP WITH TIME ZONE -- when the task failed (failure case)
);

CREATE INDEX idx_tasks_scheduled_at ON tasks (scheduled_at);
