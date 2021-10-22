package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	in = func(in In) Out {
		ch := make(Bi)
		go func() {
			defer close(ch)
			select {
			case <-done:
				return
			default:
				processInChannel(done, in, ch)
			}
		}()
		return ch
	}(in)

	for _, stage := range stages {
		in = ProcessStage(in, done, stage)
	}
	return in
}

func ProcessStage(in In, done In, stage Stage) Out {
	return func(in In) Out {
		ch := make(Bi)
		go func() {
			defer func() {
				close(ch)
				for range in { // Stages in tests can wait indefinitely with "out <- f(v)" after all the goroutines have exited.
				} // Initial `in chan` is wrapped by "managed" chan and will be closed in any case (with `in` or `done` closing).
			}()
			processInChannel(done, in, ch)
		}()
		return stage(ch)
	}(in)
}

func processInChannel(done In, in In, out Bi) {
	for {
		select {
		case v, ok := <-in:
			if !ok {
				return
			}
			select {
			case out <- v:
			case <-done:
				return
			}
		case <-done:
			return
		}
	}
}
