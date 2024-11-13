package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"hot-coffee/internal/service"
	"hot-coffee/models"
	"hot-coffee/pkg/logger"
)

type OrderHandler interface {
	CreateOrder(w http.ResponseWriter, r *http.Request)
	RetrieveOrders(w http.ResponseWriter, r *http.Request)
	RetrieveOrder(w http.ResponseWriter, r *http.Request)
	UpdateOrder(w http.ResponseWriter, r *http.Request)
	DeleteOrder(w http.ResponseWriter, r *http.Request)
	CloseOrder(w http.ResponseWriter, r *http.Request)

	WriteRawJSONResponse(statusCode int, jsonResponse any, w http.ResponseWriter, r *http.Request)
	WriteJSONResponse(statusCode int, jsonResponse any, w http.ResponseWriter, r *http.Request)
	WriteErrorResponse(statusCode int, err error, w http.ResponseWriter, r *http.Request)
}

type orderHandler struct {
	OrderService service.OrderService
	logger       *logger.Logger
}

func NewOrderHandler(s service.OrderService, l *logger.Logger) *orderHandler {
	return &orderHandler{OrderService: s, logger: l}
}

func (h *orderHandler) WriteRawJSONResponse(statusCode int, jsonResponse any, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(statusCode)

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(jsonResponse)
	if err != nil {
		h.WriteErrorResponse(http.StatusInternalServerError, err, w, r)
	}
}

func (h *orderHandler) WriteJSONResponse(statusCode int, jsonResponse any, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(statusCode)

	w.Header().Set("Content-Type", "application/json")
	formattedJSON, err := json.MarshalIndent(jsonResponse, "", " ")
	if err != nil {
		h.WriteErrorResponse(http.StatusInternalServerError, err, w, r)
		return
	}

	w.Write(formattedJSON)
}

func (h *orderHandler) WriteErrorResponse(statusCode int, err error, w http.ResponseWriter, r *http.Request) {
	// TODO: if its statusCode == 500 -> add ERROR log
	// TODO: in other cases 		  -> print DEBUG log
	// TODO: find case to add WARNING log (высосать из пальца)

	switch statusCode {
	case http.StatusInternalServerError:
		h.logger.PrintErrorMsg(err.Error())
	case http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusUnsupportedMediaType,
		http.StatusConflict:

		h.logger.PrintDebugMsg(err.Error())
	}
	errorJSON := &models.ErrorResponse{Error: err.Error()}
	h.WriteJSONResponse(statusCode, errorJSON, w, r)
}

func (h *orderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.WriteErrorResponse(http.StatusBadRequest, errors.New("request body can not be empty"), w, r)
		return
	}
	defer r.Body.Close()

	var order models.Order
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&order); err != nil {
		if err == io.EOF {
			h.WriteErrorResponse(http.StatusBadRequest, errors.New("request body can not be empty"), w, r)
			return
		}
		h.WriteErrorResponse(http.StatusBadRequest, err, w, r)
		return
	}

	h.logger.PrintDebugMsg("Creating new order: %+v", order)

	err := h.OrderService.AddOrder(order)
	if err != nil {
		switch err {
		case service.ErrNotUniqueOrder:
			h.WriteErrorResponse(http.StatusConflict, err, w, r)
			return
		default:
			h.WriteErrorResponse(http.StatusInternalServerError, err, w, r)
			return
		}
	}

	h.logger.PrintInfoMsg("Successfully created new order: %+v", order)

	h.WriteJSONResponse(http.StatusCreated, order, w, r)
}

func (h *orderHandler) RetrieveOrders(w http.ResponseWriter, r *http.Request) {
	// Retrieve the orders from the service layer
	data, err := h.OrderService.RetrieveOrders()
	if err != nil {
		switch err {
		default:
			h.WriteErrorResponse(http.StatusInternalServerError, err, w, r)
			return
		}
	}

	h.logger.PrintDebugMsg("Retrieved orders")

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *orderHandler) RetrieveOrder(w http.ResponseWriter, r *http.Request) {
	orderId := r.PathValue("id")

	if len(orderId) == 0 {
		h.WriteErrorResponse(http.StatusBadRequest, errors.New("order id is not valid"), w, r)
		return
	}

	data, err := h.OrderService.RetrieveOrder(orderId)
	if err != nil {
		if err.Error() == "order not found" {
			h.WriteErrorResponse(http.StatusNotFound, fmt.Errorf("order with id '%s' not found", orderId), w, r)
			return
		} else {
			h.WriteErrorResponse(http.StatusInternalServerError, err, w, r)
			return
		}
	}

	h.logger.PrintDebugMsg("Retrieved order with ID: %s", orderId)

	_, err = w.Write(data)
	if err != nil {
		h.WriteErrorResponse(http.StatusInternalServerError, err, w, r)
		h.logger.PrintErrorMsg("Failed to write response: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *orderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	orderId := r.PathValue("id")

	// TODO: implement logic to Update an existing order by ID.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("There will be Update an existing order by ID: " + orderId))
}

func (h *orderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	orderId := r.PathValue("id")

	if len(orderId) == 0 {
		h.WriteErrorResponse(http.StatusBadRequest, errors.New("order id is not valid"), w, r)
		return
	}

	err := h.OrderService.DeleteOrder(orderId)
	if err != nil {
		if err.Error() == "order not found" {
			h.WriteErrorResponse(http.StatusNotFound, fmt.Errorf("order with id '%s' not found", orderId), w, r)
			return
		}
	}

	h.logger.PrintDebugMsg("order with ID: %s successfully deleted", orderId)

	w.WriteHeader(http.StatusNoContent)
}

func (h *orderHandler) CloseOrder(w http.ResponseWriter, r *http.Request) {
	orderId := r.PathValue("id")

	// TODO: implement logic to Close an order by ID.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("There will be Close an order by ID: " + orderId))
}
