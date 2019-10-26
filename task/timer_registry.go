package task

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/zhao-kun/reminder-tgbot/repo"
	"github.com/zhao-kun/reminder-tgbot/telegram"
)

const (
	taskStatusRuning int = iota
	taskStatusStop
)

type (

	// Task is define a task will be periodically executed
	task struct {
		sync.Mutex

		callback CallbackFunc
		// interval
		interval time.Duration
		// name unique identy a task
		name string

		stop chan struct{}

		status int
	}
	taskRegistry struct {
		sync.Mutex
		tasks map[string]*task
	}
)

var _ Registry = &taskRegistry{}
var _ Task = &task{}

func (r *taskRegistry) StartAllTask() {
	for k := range r.tasks {
		r.runTask(r.tasks[k])
	}
}

func (r *taskRegistry) AddTask(tas Task) error {
	t, ok := tas.(*task)
	if !ok {
		return fmt.Errorf("You must add Task object initiated by NewTask factory")
	}
	r.Lock()
	defer r.Unlock()

	if _, ok := r.tasks[tas.GetName()]; ok {
		return fmt.Errorf("Task %s already register in registry", t.name)
	}
	r.tasks[tas.GetName()] = t
	return nil
}

func (r *taskRegistry) runTask(t *task) {
	t.Lock()
	defer t.Unlock()
	if t.status == taskStatusRuning {
		return
	}

	t.status = taskStatusRuning
	go func() {
		ticker := time.NewTicker(t.interval)
		setTaskStop := func() {
			t.Lock()
			defer t.Unlock()
			t.status = taskStatusStop
		}
		for t.status == taskStatusRuning {
			select {
			case <-ticker.C:
				if t.callback() == false {
					setTaskStop()
				}
			case <-t.stop:
				setTaskStop()
			}
		}
		log.Printf("Task %s was exited", t.name)
	}()
}

func (t *task) GetName() string {
	return t.name
}

// New return a Task interface
func New(name string, duration string, taskFunc CallbackFunc) (Task, error) {
	d, err := time.ParseDuration(duration)
	if err != nil {
		return nil, fmt.Errorf("Duration:%s is not valid duration representation", duration)
	}
	return &task{
		name:     name,
		interval: d,
		status:   taskStatusStop,
		callback: taskFunc,
	}, nil
}

// NewTaskRegistry return a TaskRegistry
func NewTaskRegistry() Registry {
	return &taskRegistry{
		tasks: make(map[string]*task, 10),
	}
}

// WrapWithRepoAndTelegramClient wrap function with model.Config and
// telegram.Client function to a TaskCallbackFunc
func WrapWithRepoAndTelegramClient(tgClient telegram.Client, r repo.Repo,
	c Context, f func(telegram.Client, repo.Repo, Context) bool) CallbackFunc {
	return func() bool {
		return f(tgClient, r, c)
	}
}

// NewContext return a task Context object
func NewContext() Context {
	return map[string]interface{}{}
}
