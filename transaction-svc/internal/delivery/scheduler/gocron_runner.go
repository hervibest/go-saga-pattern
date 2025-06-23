package scheduler

import (
	"context"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/utils"
	"go-saga-pattern/transaction-svc/internal/usecase"
	"strconv"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

type SchedulerRunner interface {
	Start()
}

type schedulerRunner struct {
	scheduler              gocron.Scheduler
	usecase                usecase.SchedulerUseCase
	logs                   logs.Log
	checkSchedulerDuration time.Duration
}

func NewSchedulerRunner(
	s gocron.Scheduler,
	usecase usecase.SchedulerUseCase,
	logs logs.Log,
) SchedulerRunner {
	schedulerStr := utils.GetEnv("TRANSACTION_CHECK_SCHEDULER_IN_SECONDS")
	schedulerInt, err := strconv.Atoi(schedulerStr)
	if err != nil || schedulerInt <= 0 {
		schedulerInt = 60
	}
	return &schedulerRunner{
		scheduler:              s,
		usecase:                usecase,
		logs:                   logs,
		checkSchedulerDuration: time.Duration(schedulerInt) * time.Second,
	}
}

func (r *schedulerRunner) Start() {
	jobDef := gocron.DurationJob(r.checkSchedulerDuration * time.Second)

	_, err := r.scheduler.NewJob(
		jobDef,
		gocron.NewTask(func(ctx context.Context) {
			ctx, cancel := context.WithTimeout(ctx, 4*time.Minute)
			defer cancel()

			if err := r.usecase.CheckTransactionStatus(ctx); err != nil {
				r.logs.Error("Failed to check transaction status", zap.Error(err))
			}
		}),
	)
	if err != nil {
		r.logs.Error("Failed to create job", zap.Error(err))
		return
	}
	r.logs.Info("Scheduler job created to check transaction status every 100 seconds", zap.String("job", "CheckTransactionStatus"))
	r.scheduler.Start()
}
