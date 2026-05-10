// internal/handler/product_handler.go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Bughay/go-backend-layout/internal/lib"

	"github.com/Bughay/go-backend-layout/internal/auth"
	"github.com/Bughay/go-backend-layout/internal/model"
	"github.com/Bughay/go-backend-layout/internal/service"
)

// ProductHandler handles product CRUD HTTP requests.
type ProductHandler struct {
	productSvc service.ProductService
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(productSvc service.ProductService) *ProductHandler {
	return &ProductHandler{productSvc: productSvc}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())

	var req model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.productSvc.Create(r.Context(), claims.UserID, &req)
	if err != nil {
		if isValidationErr(err) {
			lib.WriteError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		lib.WriteError(w, http.StatusInternalServerError, "failed to create product")
		return
	}

	lib.WriteJSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	result, err := h.productSvc.GetAll(r.Context(), page, limit)
	if err != nil {
		lib.WriteError(w, http.StatusInternalServerError, "failed to retrieve products")
		return
	}

	lib.WriteJSON(w, http.StatusOK, result)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Go 1.22+: PathValue extracts {id} from the route pattern
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		lib.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	product, err := h.productSvc.GetByID(r.Context(), id)
	if err != nil {
		if strings.HasPrefix(err.Error(), "not_found:") {
			lib.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		lib.WriteError(w, http.StatusInternalServerError, "failed to retrieve product")
		return
	}

	lib.WriteJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		lib.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var req model.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.productSvc.Update(r.Context(), claims.UserID, id, &req)
	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), "not_found:"):
			lib.WriteError(w, http.StatusNotFound, err.Error())
		case strings.HasPrefix(err.Error(), "forbidden:"):
			lib.WriteError(w, http.StatusForbidden, err.Error())
		case isValidationErr(err):
			lib.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			lib.WriteError(w, http.StatusInternalServerError, "failed to update product")
		}
		return
	}

	lib.WriteJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		lib.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	if err := h.productSvc.Delete(r.Context(), claims.UserID, id); err != nil {
		switch {
		case strings.HasPrefix(err.Error(), "not_found:"):
			lib.WriteError(w, http.StatusNotFound, err.Error())
		case strings.HasPrefix(err.Error(), "forbidden:"):
			lib.WriteError(w, http.StatusForbidden, err.Error())
		default:
			lib.WriteError(w, http.StatusInternalServerError, "failed to delete product")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204: Success with no body
}
