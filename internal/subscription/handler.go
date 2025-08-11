package subscription

import (
	"context"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"tz1/pkg/apperror"
	"tz1/pkg/handlers"
	"tz1/pkg/helper"
	"tz1/pkg/logging"
)

const (
	subscriptionsURL    = "/subscriptions"
	subscriptionURL     = "/subscription/:uuid"
	subscriptionsSumURL = "/subscriptions/sum"
)

type handler struct {
	logger     *logging.Logger
	repository Repository
}

func NewHandler(repository Repository, logger *logging.Logger) handlers.Handler {
	return &handler{
		repository: repository,
		logger:     logger,
	}
}

type Result struct {
	Result string `json:"result"`
}

type SumResult struct {
	Sum int64 `json:"sum"`
}

type ListResult struct {
	Result string         `json:"result"`
	List   []Subscription `json:"list"`
}

func (h *handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, subscriptionsURL, apperror.Middleware(h.GetList))
	router.HandlerFunc(http.MethodPost, subscriptionsURL, apperror.Middleware(h.Create))
	router.HandlerFunc(http.MethodGet, subscriptionURL, apperror.Middleware(h.GetOne))
	router.HandlerFunc(http.MethodPut, subscriptionURL, apperror.Middleware(h.Update))
	router.HandlerFunc(http.MethodDelete, subscriptionURL, apperror.Middleware(h.Delete))
	router.HandlerFunc(http.MethodGet, subscriptionsSumURL, apperror.Middleware(h.GetSum))
}

func (h *handler) GetList(w http.ResponseWriter, r *http.Request) error {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	user := r.URL.Query().Get("user_id")
	service := r.URL.Query().Get("service_name")
	limit := helper.GetQueryInt(r, "limit", 20)
	if limit > 1000 {
		limit = 1000
	}
	offset := helper.GetQueryInt(r, "offset", 0)
	all, err := h.repository.GetList(context.TODO(), limit, offset, from, to, user, service)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	allBytes, err := json.Marshal(all)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(allBytes)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) GetSum(w http.ResponseWriter, r *http.Request) error {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	user := r.URL.Query().Get("user_id")
	service := r.URL.Query().Get("service_name")
	sum, err := h.repository.GetSum(context.TODO(), from, to, user, service)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	sumBytes, err := json.Marshal(SumResult{Sum: sum})
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(sumBytes)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) error {
	s := Subscription{}

	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	err = h.repository.Create(context.TODO(), &s)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	sBytes, err := json.Marshal(s)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(sBytes)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) GetOne(w http.ResponseWriter, r *http.Request) error {
	id, ok := helper.UuidFromContext(r.Context())

	if !ok {
		http.NotFound(w, r)
		return nil
	}

	s, err := h.repository.FindOne(context.TODO(), id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	sBytes, err := json.Marshal(s)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(sBytes)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) Update(w http.ResponseWriter, r *http.Request) error {
	id, ok := helper.UuidFromContext(r.Context())

	if !ok {
		http.NotFound(w, r)
		return nil
	}

	s := Subscription{}

	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	err = h.repository.Update(context.TODO(), id, &s)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	sBytes, err := json.Marshal(s)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(sBytes)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, ok := helper.UuidFromContext(r.Context())

	if !ok {
		http.NotFound(w, r)
		return nil
	}

	err := h.repository.Delete(context.TODO(), id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}
