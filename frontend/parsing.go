package server

import (
	"fmt"
	"net/http"
	"strconv"
)

func getRequestInt64(req *http.Request, key string) (int64, error) {
	params := req.URL.Query()
	if strkey := params.Get(key); len(strkey) > 0 {
		val, err := strconv.ParseInt(strkey, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q => %q as int: %w", key, strkey, err)
		}
		return val, nil
	}
	return 0, nil

}

func getStoryId(req *http.Request) (int64, error) {
	return getRequestInt64(req, storyIdKey)
}

func getGameId(req *http.Request) (int64, error) {
	return getRequestInt64(req, gameIdKey)
}
