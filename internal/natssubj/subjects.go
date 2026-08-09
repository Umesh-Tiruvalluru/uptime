// Package natssubj centralizes NATS subject names so worker, scheduler, and API agree.
package natssubj

// ChecksPerform is the subject the scheduler publishes to and the worker subscribes to.
// Message body is a JSON-encoded models.CheckJob.
const ChecksPerform = "checks.perform"

// ChecksCompleted is published by the worker when a monitor check finishes.
const ChecksCompleted = "checks.completed"
