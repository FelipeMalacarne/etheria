package id

import "github.com/google/uuid"

// UUIDGenerator creates random UUIDv4 identifiers.
type UUIDGenerator struct{}

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (UUIDGenerator) New() string {
	return uuid.NewString()
}
