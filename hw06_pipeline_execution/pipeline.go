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

	curCh := in

	for _, stage := range stages {
		if stage != nil {
			stageOut := stage(curCh)
			curCh = runStage(done, stageOut)
		}
	}
	return curCh
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
