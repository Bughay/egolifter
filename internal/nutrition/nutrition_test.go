package nutrition

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Bughay/egolifter/internal/auth"
	"github.com/Bughay/egolifter/internal/lib"
)

// --- Stubs ---

// stubMealRepository records whether Create was called and returns a fixed meal.
type stubMealRepository struct {
	err          error // when set, every method fails with it
	meal         *Meal // returned by FindByID (nil = not found)
	createCalled bool
}

func (s *stubMealRepository) Create(ctx context.Context, userID string, req *CreateMealRequest) (*Meal, error) {
	s.createCalled = true
	if s.err != nil {
		return nil, s.err
	}
	return &Meal{ID: "meal-1", UserID: userID, Name: req.Name, Foods: []ConsumedFood{}}, nil
}

func (s *stubMealRepository) FindByID(ctx context.Context, userID, id string) (*Meal, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.meal, nil
}

func (s *stubMealRepository) List(ctx context.Context, userID string) ([]Meal, error) {
	if s.err != nil {
		return nil, s.err
	}
	return []Meal{{ID: "meal-1", UserID: userID, Name: "breakfast", Foods: []ConsumedFood{}}}, nil
}

// stubFoodRepository records calls and returns configurable fixtures.
type stubFoodRepository struct {
	err  error // when set, every method fails with it
	food *Food // returned by FindByID/Update (nil = not found)

	createCalled bool
	updateCalled bool
	deleteCalled bool
}

func (s *stubFoodRepository) Create(ctx context.Context, userID string, req *CreateFoodRequest) (*Food, error) {
	s.createCalled = true
	if s.err != nil {
		return nil, s.err
	}
	return &Food{
		ID:               "food-1",
		UserID:           userID,
		Name:             req.Name,
		Calories100:      req.Calories100,
		Protein100:       req.Protein100,
		Carbohydrates100: req.Carbohydrates100,
		Fat100:           req.Fat100,
	}, nil
}

func (s *stubFoodRepository) FindByID(ctx context.Context, userID, id string) (*Food, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.food, nil
}

func (s *stubFoodRepository) List(ctx context.Context, userID string) ([]Food, error) {
	if s.err != nil {
		return nil, s.err
	}
	return []Food{{ID: "food-1", Name: "oats", Calories100: 389, Protein100: 16.9}}, nil
}

func (s *stubFoodRepository) Update(ctx context.Context, userID string, req *UpdateFoodRequest) (*Food, error) {
	s.updateCalled = true
	if s.err != nil {
		return nil, s.err
	}
	return s.food, nil
}

func (s *stubFoodRepository) Delete(ctx context.Context, userID, id string) error {
	s.deleteCalled = true
	return s.err
}

// --- Meal service tests ---

func TestCreateMealValidation(t *testing.T) {
	validFoods := []MealFoodInput{
		{Name: "oats", WeightG: 150, Calories: 200, Protein: 10, Carbohydrates: 30, Fat: 5},
		{Name: "milk", WeightG: 60, Calories: 30, Protein: 2, Carbohydrates: 3, Fat: 1},
	}

	tests := []struct {
		name    string
		req     *CreateMealRequest
		wantErr string // substring of the expected error; empty means success
	}{
		{
			name: "valid request",
			req:  &CreateMealRequest{Name: "breakfast", Foods: validFoods},
		},
		{
			name:    "empty name",
			req:     &CreateMealRequest{Name: "", Foods: validFoods},
			wantErr: "meal name is required",
		},
		{
			name:    "whitespace name",
			req:     &CreateMealRequest{Name: "   ", Foods: validFoods},
			wantErr: "meal name is required",
		},
		{
			name:    "name too long",
			req:     &CreateMealRequest{Name: strings.Repeat("a", 101), Foods: validFoods},
			wantErr: "at most 100 characters",
		},
		{
			name:    "no foods",
			req:     &CreateMealRequest{Name: "breakfast", Foods: []MealFoodInput{}},
			wantErr: "at least one food",
		},
		{
			name: "blank food name",
			req: &CreateMealRequest{Name: "breakfast", Foods: []MealFoodInput{
				{Name: "  ", WeightG: 100},
			}},
			wantErr: "name is required",
		},
		{
			name: "zero weight",
			req: &CreateMealRequest{Name: "breakfast", Foods: []MealFoodInput{
				{Name: "oats", WeightG: 0},
			}},
			wantErr: "weight_g must be greater than zero",
		},
		{
			name: "negative weight",
			req: &CreateMealRequest{Name: "breakfast", Foods: []MealFoodInput{
				{Name: "oats", WeightG: -50},
			}},
			wantErr: "weight_g must be greater than zero",
		},
		{
			name: "absurd weight",
			req: &CreateMealRequest{Name: "breakfast", Foods: []MealFoodInput{
				{Name: "oats", WeightG: 5001},
			}},
			wantErr: "weight_g must be at most 5000",
		},
		{
			name: "negative macro",
			req: &CreateMealRequest{Name: "breakfast", Foods: []MealFoodInput{
				{Name: "oats", WeightG: 100, Protein: -1},
			}},
			wantErr: "macros must not be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubMealRepository{}
			svc := NewMealService(repo)

			meal, err := svc.CreateMeal(context.Background(), "user-1", tt.req)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				if !repo.createCalled {
					t.Fatal("expected repository Create to be called")
				}
				if meal == nil || meal.Name != tt.req.Name {
					t.Fatalf("unexpected meal returned: %+v", meal)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.HasPrefix(err.Error(), "validation:") {
				t.Errorf("expected validation error, got: %v", err)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tt.wantErr, err)
			}
			if repo.createCalled {
				t.Error("repository Create should not be called on validation failure")
			}
		})
	}
}

// --- Meal food resolution (catalog match / auto-save) ---

func TestResolveMealFood(t *testing.T) {
	apple := Food{ID: "f-apple", Name: "apple", Calories100: 52, Protein100: 0.3, Carbohydrates100: 14, Fat100: 0.2}
	apple2 := Food{ID: "f-apple2", Name: "apple_2", Calories100: 60, Protein100: 0.3, Carbohydrates100: 14, Fat100: 0.2}

	tests := []struct {
		name        string
		saved       []Food
		in          MealFoodInput
		wantMatchID string  // non-empty means an existing food should be reused
		wantCreate  string  // expected toCreate.Name; empty means no create expected
		wantErr     string  // substring; empty means no error
		wantCal100  float64 // checked only when checkCal is true
		checkCal    bool
	}{
		{
			name:        "exact match reuses existing food",
			saved:       []Food{apple},
			in:          MealFoodInput{Name: "apple", WeightG: 100, Calories: 52, Protein: 0.3, Carbohydrates: 14, Fat: 0.2},
			wantMatchID: "f-apple",
		},
		{
			name:       "unknown food is created",
			saved:      []Food{apple},
			in:         MealFoodInput{Name: "banana", WeightG: 100, Calories: 89, Protein: 1.1, Carbohydrates: 23, Fat: 0.3},
			wantCreate: "banana",
		},
		{
			name:       "same name different macros creates _2 variant",
			saved:      []Food{apple},
			in:         MealFoodInput{Name: "apple", WeightG: 100, Calories: 60, Protein: 0.3, Carbohydrates: 14, Fat: 0.2},
			wantCreate: "apple_2",
		},
		{
			name:        "variant exact match reuses _2",
			saved:       []Food{apple, apple2},
			in:          MealFoodInput{Name: "apple", WeightG: 100, Calories: 60, Protein: 0.3, Carbohydrates: 14, Fat: 0.2},
			wantMatchID: "f-apple2",
		},
		{
			name:    "third differing variant errors",
			saved:   []Food{apple, apple2},
			in:      MealFoodInput{Name: "apple", WeightG: 100, Calories: 70, Protein: 0.3, Carbohydrates: 14, Fat: 0.2},
			wantErr: "two foods named",
		},
		{
			name:       "totals are converted to per-100g",
			saved:      nil,
			in:         MealFoodInput{Name: "rice", WeightG: 200, Calories: 260, Protein: 5, Carbohydrates: 56, Fat: 0.6},
			wantCreate: "rice",
			wantCal100: 130,
			checkCal:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchID, toCreate, err := resolveMealFood(tt.saved, tt.in)

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantMatchID != "" {
				if matchID != tt.wantMatchID {
					t.Errorf("matchID = %q, want %q", matchID, tt.wantMatchID)
				}
				if toCreate != nil {
					t.Errorf("expected no food to create, got %+v", toCreate)
				}
				return
			}

			if toCreate == nil {
				t.Fatalf("expected a food to create (%q), got matchID=%q", tt.wantCreate, matchID)
			}
			if toCreate.Name != tt.wantCreate {
				t.Errorf("toCreate.Name = %q, want %q", toCreate.Name, tt.wantCreate)
			}
			if tt.checkCal && toCreate.Calories100 != tt.wantCal100 {
				t.Errorf("toCreate.Calories100 = %v, want %v", toCreate.Calories100, tt.wantCal100)
			}
		})
	}
}

// --- Food service tests ---

func TestCreateFoodValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateFoodRequest
		wantErr string // substring of the expected error; empty means success
	}{
		{
			name: "valid request",
			req:  &CreateFoodRequest{Name: "oats", Calories100: 389, Protein100: 16.9, Carbohydrates100: 66.3, Fat100: 6.9},
		},
		{
			name: "boundary values are valid",
			req:  &CreateFoodRequest{Name: "pure fat", Calories100: 900, Protein100: 100, Carbohydrates100: 100, Fat100: 100},
		},
		{
			name:    "empty name",
			req:     &CreateFoodRequest{Name: "", Calories100: 100},
			wantErr: "food name is required",
		},
		{
			name:    "whitespace name",
			req:     &CreateFoodRequest{Name: "   ", Calories100: 100},
			wantErr: "food name is required",
		},
		{
			name:    "name too long",
			req:     &CreateFoodRequest{Name: strings.Repeat("a", 101), Calories100: 100},
			wantErr: "at most 100 characters",
		},
		{
			name:    "negative calories",
			req:     &CreateFoodRequest{Name: "oats", Calories100: -1},
			wantErr: "calories_100 must be between 0 and 900",
		},
		{
			name:    "absurd calories",
			req:     &CreateFoodRequest{Name: "oats", Calories100: 901},
			wantErr: "calories_100 must be between 0 and 900",
		},
		{
			name:    "negative protein",
			req:     &CreateFoodRequest{Name: "oats", Calories100: 100, Protein100: -1},
			wantErr: "protein_100 must be between 0 and 100",
		},
		{
			name:    "absurd protein",
			req:     &CreateFoodRequest{Name: "oats", Calories100: 100, Protein100: 101},
			wantErr: "protein_100 must be between 0 and 100",
		},
		{
			name:    "absurd carbohydrates",
			req:     &CreateFoodRequest{Name: "oats", Calories100: 100, Carbohydrates100: 101},
			wantErr: "carbohydrates_100 must be between 0 and 100",
		},
		{
			name:    "absurd fat",
			req:     &CreateFoodRequest{Name: "oats", Calories100: 100, Fat100: 101},
			wantErr: "fat_100 must be between 0 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubFoodRepository{}
			svc := NewNutritionService(repo)

			food, err := svc.CreateFood(context.Background(), "user-1", tt.req)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				if !repo.createCalled {
					t.Fatal("expected repository Create to be called")
				}
				if food == nil || food.Name != tt.req.Name {
					t.Fatalf("unexpected food returned: %+v", food)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.HasPrefix(err.Error(), "validation:") {
				t.Errorf("expected validation error, got: %v", err)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tt.wantErr, err)
			}
			if repo.createCalled {
				t.Error("repository Create should not be called on validation failure")
			}
		})
	}
}

func TestFoodIDValidation(t *testing.T) {
	t.Run("update requires id", func(t *testing.T) {
		repo := &stubFoodRepository{}
		svc := NewNutritionService(repo)

		_, err := svc.UpdateFood(context.Background(), "user-1", &UpdateFoodRequest{ID: "  ", Name: "oats", Calories100: 100})
		if err == nil || !strings.Contains(err.Error(), "food id is required") {
			t.Fatalf("expected food id validation error, got: %v", err)
		}
		if repo.updateCalled {
			t.Error("repository Update should not be called on validation failure")
		}
	})

	t.Run("update validates payload", func(t *testing.T) {
		repo := &stubFoodRepository{}
		svc := NewNutritionService(repo)

		_, err := svc.UpdateFood(context.Background(), "user-1", &UpdateFoodRequest{ID: "food-1", Name: "", Calories100: 100})
		if err == nil || !strings.Contains(err.Error(), "food name is required") {
			t.Fatalf("expected name validation error, got: %v", err)
		}
	})

	t.Run("delete requires id", func(t *testing.T) {
		repo := &stubFoodRepository{}
		svc := NewNutritionService(repo)

		err := svc.DeleteFood(context.Background(), "user-1", "")
		if err == nil || !strings.Contains(err.Error(), "food id is required") {
			t.Fatalf("expected food id validation error, got: %v", err)
		}
		if repo.deleteCalled {
			t.Error("repository Delete should not be called on validation failure")
		}
	})

	t.Run("get requires id", func(t *testing.T) {
		repo := &stubFoodRepository{}
		svc := NewNutritionService(repo)

		_, err := svc.GetFood(context.Background(), "user-1", "  ")
		if err == nil || !strings.Contains(err.Error(), "food id is required") {
			t.Fatalf("expected food id validation error, got: %v", err)
		}
	})
}

// --- Handler test helpers ---

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func doJSON(t *testing.T, mux *http.ServeMux, method, target, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, target, reader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// newFoodServer mounts the food routes behind real JWT middleware and
// returns the mux together with a valid bearer token for user-1.
func newFoodServer(t *testing.T, repo FoodRepository) (*http.ServeMux, string) {
	t.Helper()
	mgr := auth.NewManager("test-secret", 1, 1)
	handler := NewNutritionHandler(NewNutritionService(repo), testLogger())
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, mgr.Middleware)

	token, err := mgr.Generate("user-1", "user@test.com", "user")
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}
	return mux, token
}

// newMealServer mounts the meal routes behind real JWT middleware and
// returns the mux together with a valid bearer token for user-1.
func newMealServer(t *testing.T, repo MealRepository) (*http.ServeMux, string) {
	t.Helper()
	mgr := auth.NewManager("test-secret", 1, 1)
	handler := NewMealHandler(NewMealService(repo), testLogger())
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, mgr.Middleware)

	token, err := mgr.Generate("user-1", "user@test.com", "user")
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}
	return mux, token
}

// --- Food handler tests ---

func TestFoodEndpointsRequireAuth(t *testing.T) {
	mux, _ := newFoodServer(t, &stubFoodRepository{})

	routes := []struct {
		method string
		target string
	}{
		{http.MethodPost, "/food/create"},
		{http.MethodGet, "/food/view"},
		{http.MethodPut, "/food/update"},
		{http.MethodDelete, "/food/delete"},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.target, func(t *testing.T) {
			if rec := doJSON(t, mux, rt.method, rt.target, "", nil); rec.Code != http.StatusUnauthorized {
				t.Errorf("no token: expected 401, got %d", rec.Code)
			}
			if rec := doJSON(t, mux, rt.method, rt.target, "garbage-token", nil); rec.Code != http.StatusUnauthorized {
				t.Errorf("garbage token: expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestCreateFoodEndpoint(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		mux, token := newFoodServer(t, &stubFoodRepository{})
		rec := doJSON(t, mux, http.MethodPost, "/food/create", token, CreateFoodRequest{
			Name: "oats", Calories100: 389, Protein100: 16.9, Carbohydrates100: 66.3, Fat100: 6.9,
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		var food Food
		if err := json.NewDecoder(rec.Body).Decode(&food); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if food.Name != "oats" || food.Calories100 != 389 {
			t.Errorf("unexpected food in response: %+v", food)
		}
	})

	t.Run("validation failure", func(t *testing.T) {
		mux, token := newFoodServer(t, &stubFoodRepository{})
		rec := doJSON(t, mux, http.MethodPost, "/food/create", token, CreateFoodRequest{Name: "oats", Calories100: 901})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
		}
		var apiErr lib.APIError
		if err := json.NewDecoder(rec.Body).Decode(&apiErr); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if apiErr.Code != http.StatusUnprocessableEntity || !strings.HasPrefix(apiErr.Message, "validation:") {
			t.Errorf("unexpected error body: %+v", apiErr)
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		mux, token := newFoodServer(t, &stubFoodRepository{})
		req := httptest.NewRequest(http.MethodPost, "/food/create", strings.NewReader("{not json"))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("repository failure", func(t *testing.T) {
		mux, token := newFoodServer(t, &stubFoodRepository{err: errors.New("db down")})
		rec := doJSON(t, mux, http.MethodPost, "/food/create", token, CreateFoodRequest{Name: "oats", Calories100: 389})
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}

func TestViewFoodEndpoint(t *testing.T) {
	t.Run("list all", func(t *testing.T) {
		mux, token := newFoodServer(t, &stubFoodRepository{})
		rec := doJSON(t, mux, http.MethodGet, "/food/view", token, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var foods []Food
		if err := json.NewDecoder(rec.Body).Decode(&foods); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(foods) != 1 {
			t.Errorf("expected 1 food, got %d", len(foods))
		}
	})

	t.Run("by id found", func(t *testing.T) {
		mux, token := newFoodServer(t, &stubFoodRepository{food: &Food{ID: "food-1", Name: "oats"}})
		rec := doJSON(t, mux, http.MethodGet, "/food/view?id=food-1", token, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("by id not found", func(t *testing.T) {
		mux, token := newFoodServer(t, &stubFoodRepository{food: nil})
		rec := doJSON(t, mux, http.MethodGet, "/food/view?id=missing", token, nil)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestUpdateFoodEndpoint(t *testing.T) {
	validReq := UpdateFoodRequest{ID: "food-1", Name: "oats v2", Calories100: 380}

	t.Run("found", func(t *testing.T) {
		mux, token := newFoodServer(t, &stubFoodRepository{food: &Food{ID: "food-1", Name: "oats v2"}})
		rec := doJSON(t, mux, http.MethodPut, "/food/update", token, validReq)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux, token := newFoodServer(t, &stubFoodRepository{food: nil})
		rec := doJSON(t, mux, http.MethodPut, "/food/update", token, validReq)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("missing id", func(t *testing.T) {
		mux, token := newFoodServer(t, &stubFoodRepository{})
		rec := doJSON(t, mux, http.MethodPut, "/food/update", token, UpdateFoodRequest{Name: "oats"})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestDeleteFoodEndpoint(t *testing.T) {
	t.Run("valid id", func(t *testing.T) {
		mux, token := newFoodServer(t, &stubFoodRepository{})
		rec := doJSON(t, mux, http.MethodDelete, "/food/delete?id=food-1", token, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("missing id", func(t *testing.T) {
		mux, token := newFoodServer(t, &stubFoodRepository{})
		rec := doJSON(t, mux, http.MethodDelete, "/food/delete", token, nil)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

// --- Meal handler tests ---

func TestMealEndpointsRequireAuth(t *testing.T) {
	mux, _ := newMealServer(t, &stubMealRepository{})

	routes := []struct {
		method string
		target string
	}{
		{http.MethodPost, "/meal/create"},
		{http.MethodGet, "/meal/view"},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.target, func(t *testing.T) {
			if rec := doJSON(t, mux, rt.method, rt.target, "", nil); rec.Code != http.StatusUnauthorized {
				t.Errorf("no token: expected 401, got %d", rec.Code)
			}
			if rec := doJSON(t, mux, rt.method, rt.target, "garbage-token", nil); rec.Code != http.StatusUnauthorized {
				t.Errorf("garbage token: expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestCreateMealEndpoint(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		mux, token := newMealServer(t, &stubMealRepository{})
		rec := doJSON(t, mux, http.MethodPost, "/meal/create", token, CreateMealRequest{
			Name:  "breakfast",
			Foods: []MealFoodInput{{Name: "oats", WeightG: 150, Calories: 200, Protein: 10, Carbohydrates: 30, Fat: 5}},
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		var meal Meal
		if err := json.NewDecoder(rec.Body).Decode(&meal); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if meal.Name != "breakfast" {
			t.Errorf("unexpected meal in response: %+v", meal)
		}
	})

	t.Run("validation failure", func(t *testing.T) {
		mux, token := newMealServer(t, &stubMealRepository{})
		rec := doJSON(t, mux, http.MethodPost, "/meal/create", token, CreateMealRequest{Name: ""})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestViewMealEndpoint(t *testing.T) {
	t.Run("list all", func(t *testing.T) {
		mux, token := newMealServer(t, &stubMealRepository{})
		rec := doJSON(t, mux, http.MethodGet, "/meal/view", token, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var meals []Meal
		if err := json.NewDecoder(rec.Body).Decode(&meals); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(meals) != 1 {
			t.Errorf("expected 1 meal, got %d", len(meals))
		}
	})

	t.Run("by id not found", func(t *testing.T) {
		mux, token := newMealServer(t, &stubMealRepository{meal: nil})
		rec := doJSON(t, mux, http.MethodGet, "/meal/view?id=missing", token, nil)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}
