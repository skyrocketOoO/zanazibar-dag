package usecasedom

import (
	"zanzibar-dag/domain"
)

type RelationUsecase interface {
	GetAll() ([]domain.Relation, error)
	Query(relation domain.Relation) ([]domain.Relation, error)
	Create(relation domain.Relation) error
	Delete(relation domain.Relation) error

	GetAllNamespaces() ([]string, error)
	Check(subject domain.Node, object domain.Node, searchCondition domain.SearchCondition) (bool, error)
	GetShortestPath(subject domain.Node, object domain.Node, searchCondition domain.SearchCondition) ([]domain.Relation, error)
	GetAllPaths(subject domain.Node, object domain.Node, searchCondition domain.SearchCondition) ([][]domain.Relation, error)
	GetAllObjectRelations(subject domain.Node, searchCondition domain.SearchCondition, collectCondition domain.CollectCondition) ([]domain.Relation, error)
	GetAllSubjectRelations(object domain.Node, searchCondition domain.SearchCondition, collectCondition domain.CollectCondition) ([]domain.Relation, error)

	ClearAllRelations() error
}
