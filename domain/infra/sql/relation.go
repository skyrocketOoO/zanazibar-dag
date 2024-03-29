package sqldom

import (
	"github.com/skyrocketOoO/zanazibar-dag/domain"
)

type Relation struct {
	ID               uint   `gorm:"primarykey"`
	AllColumns       string `gorm:"index"`
	ObjectNamespace  string `gorm:"index:idx_object"`
	ObjectName       string `gorm:"index:idx_object"`
	Relation         string `gorm:"index:idx_object"`
	SubjectNamespace string `gorm:"index:idx_subject"`
	SubjectName      string `gorm:"index:idx_subject"`
	SubjectRelation  string `gorm:"index:idx_subject"`
}

type RelationRepository interface {
	Create(relation domain.Relation) error
	Delete(relation domain.Relation) error
	DeleteByQueries(queries []domain.Relation) error
	BatchOperation(operations []domain.Operation) error
	GetAll(options ...PageOptions) (relations []domain.Relation, lastID uint, err error)
	Query(query domain.Relation) ([]domain.Relation, error)
	GetAllNamespaces() ([]string, error)
	DeleteAll() error
}

type PageOptions struct {
	LastID   uint
	PageSize int
}
