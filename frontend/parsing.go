package server

import (
	"fmt"
	"net/http"
	"strconv"
)

func getStoryId(req *http.Request) (int64, error) {
	params := req.URL.Query()
	if strid := params.Get(storyIdKey); len(strid) > 0 {
		sid, err := strconv.ParseInt(strid, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse story ID %q: %w", strid, err)
		}
		return sid, nil
	}
	return 0, nil
}
