package main

import (
	"context"
	"encoding/json"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/linkedin/linkedin-immediate-processor/processor"
)

// ProcessorAPI defines the interface for the processor
type ProcessorAPI interface {
	ProcessAccount(ctx context.Context, wo processor.WorkOrder) error
}

// WorkerService handles work order processing with injectable dependencies
type WorkerService struct {
	proc ProcessorAPI
	log  *logger.Logger
}

// NewWorkerService creates a new worker service
func NewWorkerService(proc ProcessorAPI, log *logger.Logger) *WorkerService {
	return &WorkerService{
		proc: proc,
		log:  log,
	}
}

// ProcessWorkOrderTestable processes a single work order
func (s *WorkerService) ProcessWorkOrderTestable(ctx context.Context, value []byte) error {
	var wo processor.WorkOrder
	if err := json.Unmarshal(value, &wo); err != nil {
		s.log.Error().Err(err).Msg("Failed to unmarshal work order")
		return err
	}

	s.log.Info().
		Str("linkedin_id", wo.AccountID).
		Str("workspace_id", wo.WorkspaceID).
		Str("sync_type", wo.SyncType).
		Msg("Processing immediate work order")

	if err := s.proc.ProcessAccount(ctx, wo); err != nil {
		s.log.Error().
			Err(err).
			Str("linkedin_id", wo.AccountID).
			Str("workspace_id", wo.WorkspaceID).
			Msg("Failed to process account")
		return err
	}

	s.log.Info().
		Str("linkedin_id", wo.AccountID).
		Str("workspace_id", wo.WorkspaceID).
		Msg("Successfully processed account")

	return nil
}

// WorkerLoopTestable processes work orders from a channel
func (s *WorkerService) WorkerLoopTestable(ctx context.Context, workChan <-chan workMessage) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-workChan:
			if !ok {
				return
			}
			_ = s.ProcessWorkOrderTestable(msg.ctx, msg.value)
		}
	}
}

// ParseWorkOrderTestable parses a work order from JSON
func ParseWorkOrderTestable(value []byte) (*processor.WorkOrder, error) {
	var wo processor.WorkOrder
	if err := json.Unmarshal(value, &wo); err != nil {
		return nil, err
	}
	return &wo, nil
}

// ValidateWorkOrderTestable validates a work order
func ValidateWorkOrderTestable(wo *processor.WorkOrder) bool {
	if wo.ID == "" {
		return false
	}
	if wo.AccountID == "" {
		return false
	}
	if wo.WorkspaceID == "" {
		return false
	}
	return true
}
