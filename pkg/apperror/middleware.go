package apperror

import (
	"errors"
	"net/http"
)

type appHandler func(w http.ResponseWriter, r *http.Request) error

func Middleware(h appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		err := h(w, r)
		if err != nil {
			var appErr *AppError
			if errors.As(err, &appErr) {
				if errors.Is(err, ErrNotFound) {
					http.Error(w, string(ErrNotFound.Marshal()), http.StatusNotFound)
					return
				}

				http.Error(w, string(appErr.Marshal()), http.StatusBadRequest)
				return
			}

			http.Error(w, string(systemError(err).Marshal()), http.StatusInternalServerError)
		}
	}
}
