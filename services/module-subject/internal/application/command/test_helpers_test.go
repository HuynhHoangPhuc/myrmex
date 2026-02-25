package command

import (
	"github.com/google/uuid"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
)

// newTestSubject creates a valid Subject for use in tests.
func newTestSubject(code, name string, credits int32) *entity.Subject {
	return &entity.Subject{
		ID:       uuid.New(),
		Code:     code,
		Name:     name,
		Credits:  credits,
		IsActive: true,
	}
}
