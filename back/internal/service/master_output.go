package service

import "github.com/aleksandrboro/slotvio/internal/domain"

type PublicProfileOutput struct {
	Master   domain.Master
	Services []domain.Service
}
