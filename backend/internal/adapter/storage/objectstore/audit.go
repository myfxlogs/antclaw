package objectstore

import (
"context"
"errors"
)

// ErrNotImplemented indicates the audit store is not yet implemented.
// This is a placeholder for P6b WORM implementation.
var ErrNotImplemented = errors.New("audit WORM store not implemented")

// AuditStore is a placeholder for MinIO WORM audit storage.
// P6b: Skeleton only, full implementation in later phase.
type AuditStore struct {
enabled bool
}

// NewAuditStore creates an audit store placeholder.
func NewAuditStore() *AuditStore {
return &AuditStore{enabled: false}
}

// WriteAuditLog placeholder.
func (as *AuditStore) WriteAuditLog(ctx context.Context, auditID int64, data []byte) error {
return ErrNotImplemented
}
