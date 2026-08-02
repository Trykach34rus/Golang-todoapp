package domain

import "time"

type Statistics struct {
	TasksCreated              int
	TasksCompleted            int
	TasksCompletedRate        *float64
	TaskAverageCompletionTime *time.Duration
}

func NewStatistics(
	tasksCreated int,
	tasksCompleted int,
	tasksCompletedRate *float64,
	taskAverageCompletionTime *time.Duration,
) Statistics {
	return Statistics{
		TasksCreated: tasksCreated,
		TasksCompleted: tasksCompleted,
		TasksCompletedRate: tasksCompletedRate,
		TaskAverageCompletionTime: taskAverageCompletionTime,
	}
}