package task

type (
	// Context is context will passed between task callback func
	Context map[string]interface{}

	// CallbackFunc is task callback func
	CallbackFunc func() bool

	// Registry manager task
	Registry interface {
		StartAllTask()
		AddTask(Task) error
	}
	// Task represent a task
	Task interface {
		GetName() string
	}
)
