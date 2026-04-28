package notify

import "context"

type Service struct{}
func NewService() *Service { return &Service{} }
func (s *Service) SendNotification(ctx context.Context, userID, message string) error { return nil }
