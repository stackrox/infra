package scheduler

import (
	"fmt"
	"github.com/stackrox/rox/pkg/concurrency"
	"reflect"
	"sync"
	"time"
)

type schedulerImpl struct {
	interruptC chan struct{}
	stopSig concurrency.Signal
	stoppedSig concurrency.Signal
	newjobs []*job
	jobs []*job
	newJobMutex sync.Mutex

	wakeUpInterval time.Duration
	jobsInParallel int
}

type SchedulerOption func(*schedulerImpl) error

func SetWakeUpInterval(t time.Duration) SchedulerOption {
	return func(s *schedulerImpl) error {
		s.wakeUpInterval = t
		return nil
	}
}

func SetJobsInParallel(p int) SchedulerOption {
	return func (s *schedulerImpl) error {
		s.jobsInParallel = p
		return nil
	}
}

func (s *schedulerImpl) Start() {
	s.stoppedSig.Reset()
	go func() {
		defer s.stoppedSig.Signal()
		for !s.stopSig.IsDone() {
			select {
			case <-s.stopSig.Done():
				fmt.Println("Scheduler is shutting down.")
				close(s.interruptC)
				return
			case <-s.interruptC:
				s.executeJobs()
			case <-time.After(s.wakeUpInterval):
				s.executeJobs()
			}
		}
	}()

	fmt.Println("Starting scheduler...")
}

func (s *schedulerImpl) Stop() concurrency.Waitable {
	s.stopSig.Signal()
	return &s.stoppedSig
}

func (s *schedulerImpl) Interrupt() bool {
	select {
	case s.interruptC <- struct{}{}:
		return true
	case <-s.stoppedSig.Done():
		return false
	default:
		// If the above two cases block, we are not stopped and could not write to the channel. Since the channel is
		// buffered, there already is an interrupt pending, so no need for an additional one.
		return true
	}
}

func (s *schedulerImpl) ExecuteEvery(duration time.Duration, execFunction interface{}, args ...interface{}) error {
	if job, err := getJob(duration, execFunction, args...); err != nil {
		return err
	} else {
		job.repeat = true
		s.newJobMutex.Lock()
		defer s.newJobMutex.Unlock()
		s.newjobs = append(s.newjobs, job)
	}

	return nil
}

func (s *schedulerImpl) ExecuteOnceAfter(duration time.Duration, execFunction interface{}, args ...interface{}) error {
	if job, err := getJob(duration, execFunction, args...); err != nil {
		return err
	} else {
		s.newJobMutex.Lock()
		defer s.newJobMutex.Unlock()
		s.newjobs = append(s.newjobs, job)
	}

	return nil
}

func getJob(duration time.Duration, execFunction interface{}, args ...interface{}) (*job, error) {
	f := reflect.ValueOf(execFunction)
	if f.Kind() != reflect.Func {
		return nil, fmt.Errorf("Job should be a function.")
	}

	params := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		params[i] = reflect.ValueOf(args[i])
	}

	return &job{
		execFunction: f,
		nextRun: time.Now().Add(duration),
		duration: duration,
		params: params,
	}, nil
}

func (s *schedulerImpl) executeJobs() {
	s.newJobMutex.Lock()
	s.jobs = append(s.jobs, s.newjobs...)
	s.newjobs = nil
	s.newJobMutex.Unlock()

	pruneJobIdxs := make(map[int]struct{})
	pChan := make(chan struct{}, s.jobsInParallel)
	var wg sync.WaitGroup
	for idx, j := range s.jobs {
		if time.Now().Before(j.nextRun) {
			continue
		}

		pChan <- struct{}{}
		wg.Add(1)
		go func(idx int, j *job) {
			defer func() {
				wg.Done()
				<- pChan
			}()

			j.Execute()
			if !j.repeat {
				pruneJobIdxs[idx] = struct{}{}
			}
		}(idx, j)
	}

	wg.Wait()
	// prune jobs
	cleanedJobsList := make([]*job, len(s.jobs) - len(pruneJobIdxs))
	pruned := 0
	for i := 0; i < len(s.jobs); i++ {
		if _, ok := pruneJobIdxs[i]; ok {
			pruned++
			continue
		}

		cleanedJobsList[i-pruned] = s.jobs[i]
	}

	s.jobs = cleanedJobsList
}

type job struct {
	nextRun time.Time
	duration time.Duration
	execFunction reflect.Value
	params []reflect.Value
	repeat bool
}

func (j *job) Execute() {
	j.execFunction.Call(j.params)
	if j.repeat {
		j.nextRun = time.Now().Add(j.duration)
	}
}