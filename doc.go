/*
The workerpool package is used to run Jobs with a limited number of workers.

# Overview

(This page describes the nature of the individual package.)
More information at https://github.com/roncewind/workerpool

# Job interface

To use the workerpool, a channel of objects adhering to the Job interface must be
created.  The Job interface has two methods:

	type Job interface {
		Execute() error
		OnError(error)
	}

Job.Execute() is called for each Job in the channel, if an error is returned from
Execute(), then OnError(error) is called with the error.  If Job.Execute() panics,
OnError(error) is also called with the error constructed based on the nature of
the panic.  If panic is called with an error type, then that error is given to
OnError(error).  If panic is called with a string type, then an error is created
from that string and passed to OnError(error).  If panic is called with any other
type, a static, predefined error(ErrPanic) is returned.

# Worker pool

A new workerpool is created with the NewWorkerPool(workerCount int, jobQ chan Job)
function.  This function returns an object that conforms to the WorkerPool
interface.  To start the workers on Executing() Jobs, call workerpool.Start(context).
The given context is passed to each worker and allows the calling code the ability
to cancel all the workers once the current Job each worker is working on has
finished executing.

# Examples

Examples of use can be seen in the xxxx_test.go files.
*/
package workerpool
