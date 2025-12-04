package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if len(stages) == 0 {
		return in
	}

	for _, stage := range stages {
		if stage != nil {
			in = runStage(done, stage(in))
		}
	}
	return in
}

// Функция обработки для стейджа.
func runStage(done In, in In) Out {
	outCh := make(Bi)

	go func() {
		defer func() {
			close(outCh)
			//nolint:all
			for range in {
				// для TestAllStageStop/done_case
			}
		}()
		for {
			select {
			case <-done:
				return
			case val, ok := <-in:
				if !ok {
					return
				}
				outCh <- val
			}
		}
	}()

	return outCh
}
