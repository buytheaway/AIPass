package telemetry

import "context"

func ShutdownNoop(context.Context) error {
	return nil
}
