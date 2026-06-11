package training

import (
	"context"
	"time"
)

// TrainingService defines the contract for training business logic.
type TrainingService interface {
	SaveRoutine(ctx context.Context, userID string, req *CreateRoutineRequest) (*Routine, error)
	ListRoutines(ctx context.Context, userID string) ([]Routine, error)
	LogRoutine(ctx context.Context, userID string, req *LogWorkoutRequest) (*Workout, error)
	ListWorkoutsByDate(ctx context.Context, userID, dateStr string) ([]Workout, error)
}

type trainingService struct {
	trainingRepo TrainingRepository
}

// NewTrainingService creates a new TrainingService.
func NewTrainingService(trainingRepo TrainingRepository) TrainingService {
	return &trainingService{trainingRepo: trainingRepo}
}

func (s *trainingService) SaveRoutine(ctx context.Context, userID string, req *CreateRoutineRequest) (*Routine, error) {
	if err := validateRoutineInput(req.Name, req.Entries); err != nil {
		return nil, err
	}
	return s.trainingRepo.CreateRoutine(ctx, userID, req)
}

func (s *trainingService) ListRoutines(ctx context.Context, userID string) ([]Routine, error) {
	return s.trainingRepo.ListRoutines(ctx, userID)
}

// LogRoutine snapshots the routine's entries into a new performed workout.
// Returns (nil, nil) when the routine does not exist or belongs to another user.
func (s *trainingService) LogRoutine(ctx context.Context, userID string, req *LogWorkoutRequest) (*Workout, error) {
	if err := validateLogRequest(req.RoutineID); err != nil {
		return nil, err
	}
	routine, err := s.trainingRepo.FindRoutineByID(ctx, userID, req.RoutineID)
	if err != nil {
		return nil, err
	}
	if routine == nil {
		return nil, nil
	}
	return s.trainingRepo.LogWorkout(ctx, userID, routine.Name, routine.Entries)
}

// ListWorkoutsByDate lists workouts performed on the given YYYY-MM-DD date,
// defaulting to today when dateStr is empty.
func (s *trainingService) ListWorkoutsByDate(ctx context.Context, userID, dateStr string) ([]Workout, error) {
	date := time.Now()
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, err
		}
		date = parsed
	}
	return s.trainingRepo.ListWorkoutsByDate(ctx, userID, date)
}
