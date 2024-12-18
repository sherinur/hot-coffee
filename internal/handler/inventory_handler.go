package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"hot-coffee/internal/service"
	"hot-coffee/internal/utils"
	"hot-coffee/models"
	"hot-coffee/pkg/logger"
)

type InventoryHandler interface {
	AddInventoryItem(w http.ResponseWriter, r *http.Request)
	GetInventoryItems(w http.ResponseWriter, r *http.Request)
	GetInventoryItem(w http.ResponseWriter, r *http.Request)
	UpdateInventoryItem(w http.ResponseWriter, r *http.Request)
	DeleteInventoryItem(w http.ResponseWriter, r *http.Request)
}

type inventoryHandler struct {
	InventoryService service.InventoryService
	logger           *logger.Logger
}

func NewInventoryHandler(s service.InventoryService, l *logger.Logger) *inventoryHandler {
	return &inventoryHandler{InventoryService: s, logger: l}
}

// AddInventoryItem handles the HTTP request to add a new inventory item.
// It processes the incoming request, validates the input, and interacts with the service layer to add the item.
// If successful, it returns the added item as a JSON response with a 201 status code.
func (h *inventoryHandler) AddInventoryItem(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		utils.WriteErrorResponse(http.StatusBadRequest, errors.New("request body can not be empty"), w, r)
		return
	}
	defer r.Body.Close()

	var item models.InventoryItem

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		if err == io.EOF {
			utils.WriteErrorResponse(http.StatusBadRequest, errors.New("request body can not be empty"), w, r)
			return
		}
		utils.WriteErrorResponse(http.StatusBadRequest, err, w, r)
		return
	}

	err := h.InventoryService.AddInventoryItem(item)
	if err != nil {
		switch err {
		case service.ErrNotUniqueID:
			utils.WriteErrorResponse(http.StatusConflict, err, w, r)
			return
		case service.ErrNotValidIngredientID, service.ErrNotValidIngredientName, service.ErrNotValidQuantity, service.ErrNotValidUnit:
			utils.WriteErrorResponse(http.StatusBadRequest, err, w, r)
			return
		default:
			utils.WriteErrorResponse(http.StatusInternalServerError, err, w, r)
			return
		}
	}

	h.logger.PrintDebugMsg("Adding new inventory item: %+v", item)
	h.logger.PrintInfoMsg("Successfully added new inventory item: %+v", item)

	utils.WriteJSONResponse(http.StatusCreated, item, w, r)
}

// 200 OK — запрос был успешно обработан.
// 201 Created — новый ресурс был успешно создан.
// 400 Bad Request — ошибка в запросе.
// 500 Internal Server Error — ошибка на сервере.

// GetInventoryItems handles the HTTP request to retrieve inventory items.
// It calls the service layer to get the list of inventory items, handles errors, and returns the data in the response.
func (h *inventoryHandler) GetInventoryItems(w http.ResponseWriter, r *http.Request) {
	data, err := h.InventoryService.RetrieveInventoryItems()
	if err != nil {
		switch err {
		default:
			utils.WriteErrorResponse(http.StatusInternalServerError, err, w, r)
			return
		}
	}

	h.logger.PrintDebugMsg("Retrieved inventory items")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetInventoryItem handles the HTTP request to retrieve a specific inventory item by its ID.
func (h *inventoryHandler) GetInventoryItem(w http.ResponseWriter, r *http.Request) {
	itemId := r.PathValue("id")

	if len(itemId) == 0 {
		utils.WriteErrorResponse(http.StatusBadRequest, errors.New("identificator is not valid"), w, r)
		return
	}

	data, err := h.InventoryService.RetrieveInventoryItem(itemId)
	if err != nil {
		switch err.Error() {
		case "item not found":
			utils.WriteErrorResponse(http.StatusNotFound, fmt.Errorf("item with id '%s' not found", itemId), w, r)
		default:
			utils.WriteErrorResponse(http.StatusInternalServerError, err, w, r)
		}
		return
	}

	h.logger.PrintDebugMsg("Retrieved inventory item with ID: %s", itemId)

	// Send an HTTP status code 200 (OK) and write the retrieved item data to the response body.
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		h.logger.PrintErrorMsg("Failed to write response: %v", err)

		utils.WriteErrorResponse(http.StatusInternalServerError, err, w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdateInventoryItem handles the HTTP request to update an existing inventory item by its ID.
func (h *inventoryHandler) UpdateInventoryItem(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		utils.WriteErrorResponse(http.StatusBadRequest, errors.New("request body can not be empty"), w, r)
		return
	}
	defer r.Body.Close()

	itemId := r.PathValue("id")
	if len(itemId) == 0 {
		utils.WriteErrorResponse(http.StatusBadRequest, errors.New("item id is not valid"), w, r)
		return
	}

	var item models.InventoryItem
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		// If the request body cannot be decoded, return a Bad Request (400) response.
		if err == io.EOF {
			utils.WriteErrorResponse(http.StatusBadRequest, errors.New("request body can not be empty"), w, r)
			return
		}
		utils.WriteErrorResponse(http.StatusBadRequest, err, w, r)
		return
	}

	err := h.InventoryService.UpdateInventoryItem(itemId, item)
	if err != nil {
		switch err {
		case service.ErrNoItem:
			utils.WriteErrorResponse(http.StatusNotFound, fmt.Errorf("item with id '%s' not found", itemId), w, r)
			return
		case service.ErrNotUniqueID,
			service.ErrNotValidIngredientID,
			service.ErrNotValidIngredientName,
			service.ErrNotValidQuantity,
			service.ErrNotValidUnit:
			utils.WriteErrorResponse(http.StatusBadRequest, err, w, r)
			return
		default:
			utils.WriteErrorResponse(http.StatusInternalServerError, err, w, r)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *inventoryHandler) DeleteInventoryItem(w http.ResponseWriter, r *http.Request) {
	itemId := r.PathValue("id")
	if len(itemId) == 0 {
		utils.WriteErrorResponse(http.StatusBadRequest, errors.New("item id is not valid"), w, r)
		return
	}

	err := h.InventoryService.DeleteInventoryItem(itemId)
	if err != nil {
		switch err {
		case service.ErrNoItem:
			utils.WriteErrorResponse(http.StatusNotFound, fmt.Errorf("item with id '%s' not found", itemId), w, r)
			return
		default:
			utils.WriteErrorResponse(http.StatusInternalServerError, err, w, r)
			return
		}
	}

	h.logger.PrintDebugMsg("Inventory item with ID: %s successfully deleted", itemId)

	w.WriteHeader(http.StatusNoContent)
}
