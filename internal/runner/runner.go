package runner

import (
		"context"
		"sync"
		"github.com/Dani-Bhai/dev-doctor/internal/model"
)

type Check interface {
		Name() string 
		Run(ctx context.Context) model.Result
}

func Run(ctx context.Context, checks []Check) []model.Result {
		results := make([]model.Result, len(checks))

		var wg sync.WaitGroup

		for i, check := range checks {
				wg.Add(1)

				go func(index int, c Check){
						defer wg.Done()

						results[index] = c.Run(ctx)
				}(i, check)
		}

		wg.Wait()

		return results
} 
