package http

import (
	"net/http"
	"strconv"
)

func parsePagination(r *http.Request) (uint64, uint64) {
	query := r.URL.Query()

	limit, err := strconv.ParseUint(query.Get("limit"), 10, 64)
	if err != nil || limit == 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.ParseUint(query.Get("offset"), 10, 64)
	if err != nil {
		offset = 0
	}

	return limit, offset
}
