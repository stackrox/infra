package scheduler

import (
	"github.com/stackrox/rox/pkg/concurrency"
	"time"
)

const (
	wakeUpInterval = 1*time.Second
	jobsInParallel = 4
)

type Scheduler interface {
	Start()
	Stop()
	Interrupt()
	ExecuteEvery(duration time.Duration, execFunction interface{}, args ...interface{}) error
	ExecuteOnceAfter(duration time.Duration, execFunction interface{}, args ...interface{}) error
}

func NewScheduler(opts ...SchedulerOption) *schedulerImpl {
	scheduler := &schedulerImpl{
		interruptC: make(chan struct{}, 1),
		stopSig: concurrency.NewSignal(),
		stoppedSig: concurrency.NewSignal(),
		wakeUpInterval: wakeUpInterval,
		jobsInParallel: jobsInParallel,
	}

	applyOptions(scheduler, opts...)
	return scheduler
}

func applyOptions(scheduler *schedulerImpl, opts ...SchedulerOption) error {
	for _, opt := range opts {
		if err := opt(scheduler); err != nil {
			return err
		}
	}

	return nil
}