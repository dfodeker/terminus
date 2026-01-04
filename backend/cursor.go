package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type PageParams struct {
	Limit  int
	Cursor string // opaque to clients
}

type PageInfo struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

type PageResponse[T any] struct {
	Data []T      `json:"data"`
	Page PageInfo `json:"page"`
}

func ParsePageParams(r *http.Request, defaultLimit, maxLimit int) (PageParams, error) {
	q := r.URL.Query()
	limit := defaultLimit
	if s := q.Get("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 {
			return PageParams{}, errors.New("invalid limit")
		}
		if v > maxLimit {
			v = maxLimit
		}
		limit = v
	}
	return PageParams{
		Limit:  limit,
		Cursor: q.Get("cursor"),
	}, nil
}

type CursorCodec[T any] struct {
	// Validate is optional but recommended.
	// If set, it should return nil only for a valid decoded cursor.
	Validate func(T) error
}

func (c CursorCodec[T]) Encode(v T) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (c CursorCodec[T]) Decode(s string) (v T, ok bool, err error) {
	if s == "" {
		var zero T
		return zero, false, nil
	}

	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return v, false, errors.New("invalid cursor")
	}

	if err := json.Unmarshal(b, &v); err != nil {
		return v, false, errors.New("invalid cursor")
	}

	if c.Validate != nil {
		if err := c.Validate(v); err != nil {
			return v, false, errors.New("invalid cursor")
		}
	}

	return v, true, nil
}

type userCursor struct {
	CreatedAt string `json:"created_at"`
	ID        string `json:"id"`
}

func encodeCursor(c userCursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodeCursor(s string) (userCursor, error) {
	if s == "" {
		return userCursor{}, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return userCursor{}, errors.New("invalid cursor")
	}
	var c userCursor
	if err := json.Unmarshal(b, &c); err != nil {
		return userCursor{}, errors.New("invalid cursor")
	}
	return c, nil
}
