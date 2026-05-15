// Package postgres provides social cursor encode/decode shared by Feed and Trader repositories.
package postgres

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// SocialCursor carries the composite cursor fields used for stable pagination
// across Feed (created_at DESC, id DESC) and Comments (created_at ASC, id ASC).
type SocialCursor struct {
	CreatedAt time.Time `json:"ca"`
	ID        string    `json:"id"`
}

// EncodeSocialCursor returns a base64url-encoded cursor string.
// If cursor is nil or its ID is empty, returns "".
func EncodeSocialCursor(c *SocialCursor) string {
	if c == nil || c.ID == "" {
		return ""
	}
	b, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeSocialCursor parses a base64url-encoded cursor string.
// Returns nil if the string is empty or cannot be decoded.
func DecodeSocialCursor(raw string) (*SocialCursor, error) {
	if raw == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	var c SocialCursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if c.ID == "" {
		return nil, nil
	}
	return &c, nil
}

// CursorDirection specifies pagination direction.
type CursorDirection int

const (
	CursorDesc CursorDirection = iota // (field) < (cursor)
	CursorAsc                         // (field) > (cursor)
)

// AppendCursor adds a composite cursor clause to a SQL query.
//
//	expr: the field expression to compare, e.g. "p.created_at, p.id"
//	dir: CursorDesc or CursorAsc
//	base: the 1-based parameter index to start from (e.g. 3 means $3, $4)
func AppendCursor(query string, args []interface{}, cursor *SocialCursor, expr string, dir CursorDirection, base int) (string, []interface{}) {
	if cursor == nil || cursor.ID == "" {
		return query, args
	}
	var op string
	if dir == CursorAsc {
		op = ">"
	} else {
		op = "<"
	}
	return query + fmt.Sprintf(` AND (%s) %s ($%d, $%d)`, expr, op, base, base+1),
		append(args, cursor.CreatedAt, cursor.ID)
}
