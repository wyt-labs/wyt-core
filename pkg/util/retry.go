package util

import (
	"time"
)

func Retry(retryInterval time.Duration, retryTime int, worker func() (needRetry bool, err error)) error {
	needRetry, err := worker()
	if !needRetry {
		return err
	}

	timer := time.NewTicker(retryInterval)
	defer timer.Stop()
	totalTime := 1
	for {
		<-timer.C
		needRetry, err = worker()
		if !needRetry {
			return err
		}

		totalTime++
		if totalTime >= retryTime {
			return err
		}
	}
}

func BackoffRetry(retryTime int, retryBaseInterval time.Duration, backoffFactor int, worker func() (needRetry bool, err error)) error {
	interval := retryBaseInterval
	var needRetry bool
	var err error
	for i := 0; i < retryTime; i++ {
		needRetry, err = worker()
		if !needRetry {
			return err
		}
		time.Sleep(interval)
		interval = time.Duration(backoffFactor) * interval
	}
	return err
}
