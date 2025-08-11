package helper

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func DoWithTries(fn func() error, attemtps int, delay time.Duration) (err error) {
	for attemtps > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			attemtps--
			continue
		}
		return nil
	}
	return
}

func UuidFromContext(ctx context.Context) (string, bool) {
	ps := ctx.Value(httprouter.ParamsKey)
	switch t := ps.(type) {
	default:
		fmt.Printf("Type of httprouter.Params is %s", t)
		return "", false
	case httprouter.Params:
		uuid := ps.(httprouter.Params).ByName("uuid")
		return uuid, IsValidUUID(uuid)
	}
}

func IsValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

func FormatQuery(q string) string {
	return strings.ReplaceAll(strings.ReplaceAll(q, "\t", ""), "\n", " ")
}

func ParsePgDate(dateStr string) (pgtype.Date, error) {
	var pgDate pgtype.Date
	t, err := ParseDate(dateStr)
	if err != nil {
		pgDate.Valid = false
		return pgDate, err
	}
	pgDate.Valid = true
	pgDate.Time = t
	return pgDate, nil
}

func ParseDate(dateStr string) (time.Time, error) {
	parts := strings.Split(dateStr, "-")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid date format")
	}
	return time.Parse("2006-01", fmt.Sprintf("%s-%s", parts[1], parts[0]))
}

func GetQueryInt(r *http.Request, key string, defaultValue int) int {
	if valueStr := r.URL.Query().Get(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}
