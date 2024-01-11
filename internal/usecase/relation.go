package usecase

import (
	"fmt"
	"os"
	"time"

	"github.com/skyrocketOoO/go-utility/queue"
	"github.com/skyrocketOoO/go-utility/set"
	"github.com/skyrocketOoO/zanazibar-dag/domain"
	sqldomain "github.com/skyrocketOoO/zanazibar-dag/domain/infra/sql"
	"github.com/skyrocketOoO/zanazibar-dag/utils"
	"gorm.io/gorm"
)

type RelationUsecase struct {
	RelationRepo sqldomain.RelationRepository
}

func NewRelationUsecase(relationRepo sqldomain.RelationRepository) *RelationUsecase {
	return &RelationUsecase{
		RelationRepo: relationRepo,
	}
}

func (u *RelationUsecase) Get(relation domain.Relation) ([]domain.Relation, error) {
	startTime := time.Now()
	if relation.ObjectNamespace == "" && relation.ObjectName == "" && relation.Relation == "" &&
		relation.SubjectNamespace == "" && relation.SubjectName == "" && relation.SubjectRelation == "" {
		relations, err := u.RelationRepo.GetAll()
		if err != nil {
			return nil, err
		}
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		file, err := os.OpenFile("durations.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return relations, err
		}
		_, err = fmt.Fprintf(file, "%s %v\n", startTime.String(), duration)
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
		defer file.Close()
		return relations, nil
	} else {
		return u.RelationRepo.Query(relation)
	}
}

func (u *RelationUsecase) Create(relation domain.Relation, existOk bool) error {
	if err := utils.ValidateRelation(relation); err != nil {
		return err
	}
	ok, err := u.Check(
		domain.Node{
			Namespace: relation.ObjectNamespace,
			Name:      relation.ObjectName,
			Relation:  relation.Relation,
		},
		domain.Node{
			Namespace: relation.SubjectNamespace,
			Name:      relation.SubjectName,
			Relation:  relation.SubjectRelation,
		},
		domain.SearchCondition{},
	)
	if err != nil {
		return err
	}
	if ok {
		return domain.CauseCycleError{}
	}

	err = u.RelationRepo.Create(relation)
	if err != nil {
		if err == gorm.ErrDuplicatedKey {
			if existOk {
				return nil
			} else {
				return err
			}
		}
		return err
	}
	return nil
}

func (u *RelationUsecase) Delete(relation domain.Relation) error {
	if err := utils.ValidateRelation(relation); err != nil {
		return err
	}
	return u.RelationRepo.Delete(relation)
}

func (u *RelationUsecase) DeleteByQueries(queries []domain.Relation) error {
	return u.RelationRepo.DeleteByQueries(queries)
}

func (u *RelationUsecase) BatchOperation(operations []domain.Operation) error {
	for _, operation := range operations {
		if err := utils.ValidateRelation(operation.Relation); err != nil {
			return err
		}
	}
	return u.RelationRepo.BatchOperation(operations)
}

func (u *RelationUsecase) GetAllNamespaces() ([]string, error) {
	return u.RelationRepo.GetAllNamespaces()
}

func (u *RelationUsecase) Check(subject domain.Node, object domain.Node, searchCondition domain.SearchCondition) (bool, error) {
	if err := utils.ValidateNode(object, false); err != nil {
		return false, err
	}
	if err := utils.ValidateNode(subject, true); err != nil {
		return false, err
	}
	visited := set.NewSet[domain.Node]()
	q := queue.NewQueue[domain.Node]()
	visited.Add(subject)
	q.Push(subject)

	for !q.IsEmpty() {
		qLen := q.Len()
		for i := 0; i < qLen; i++ {
			node, _ := q.Pop()
			query := domain.Relation{
				SubjectNamespace: node.Namespace,
				SubjectName:      node.Name,
				SubjectRelation:  node.Relation,
			}
			tuples, err := u.RelationRepo.Query(query)
			if err != nil {
				return false, err
			}

			for _, tuple := range tuples {
				if tuple.ObjectNamespace == object.Namespace && tuple.ObjectName == object.Name && tuple.Relation == object.Relation {
					return true, nil
				}
				child := domain.Node{
					Namespace: tuple.ObjectNamespace,
					Name:      tuple.ObjectName,
					Relation:  tuple.Relation,
				}
				if !searchCondition.ShouldStop(child) && !visited.Exist(child) {
					visited.Add(child)
					q.Push(child)
				}
			}
		}
	}

	return false, nil
}

func (u *RelationUsecase) GetShortestPath(subject domain.Node, object domain.Node, searchCondition domain.SearchCondition) ([]domain.Relation, error) {
	if err := utils.ValidateNode(object, false); err != nil {
		return nil, err
	}
	if err := utils.ValidateNode(subject, true); err != nil {
		return nil, err
	}
	visited := set.NewSet[domain.Node]()
	type NodeItem struct {
		Cur  domain.Node
		Path []domain.Relation
	}
	firstNode := NodeItem{
		Cur:  subject,
		Path: []domain.Relation{},
	}
	q := queue.NewQueue[NodeItem]()
	visited.Add(subject)
	q.Push(firstNode)
	for !q.IsEmpty() {
		qLen := q.Len()
		for i := 0; i < qLen; i++ {
			node, _ := q.Pop()
			query := domain.Relation{
				SubjectNamespace: node.Cur.Namespace,
				SubjectName:      node.Cur.Name,
				SubjectRelation:  node.Cur.Relation,
			}
			tuples, err := u.RelationRepo.Query(query)
			if err != nil {
				return nil, err
			}

			for _, tuple := range tuples {
				if tuple.ObjectNamespace == object.Namespace && tuple.ObjectName == object.Name && tuple.Relation == object.Relation {
					return append(node.Path, tuple), nil
				}
				child := domain.Node{
					Namespace: tuple.ObjectNamespace,
					Name:      tuple.ObjectName,
					Relation:  tuple.Relation,
				}
				if !searchCondition.ShouldStop(child) && !visited.Exist(child) {
					visited.Add(child)
					copyPath := append(node.Path, tuple)
					q.Push(NodeItem{
						Cur:  child,
						Path: copyPath,
					})
				}
			}
		}
	}

	return nil, nil

	// maxDepth, err := strconv.Atoi(viper.GetString("main.max-search-depth"))
	// if err != nil {
	// 	return nil, err
	// }
	// visited := set.NewSet[domain.Node]()
	// type NodeItem struct {
	// 	Cur  domain.Node
	// 	Path []domain.Relation
	// }
	// finalPath := []domain.Relation{}
	// var dfsErr error
	// var dfs func(subject, object domain.Node, path *[]domain.Relation, cur_depth int)
	// dfs = func(subject, object domain.Node, path *[]domain.Relation, cur_depth int) {
	// 	if visited.Exist(subject) || cur_depth >= maxDepth || len(finalPath) != 0 || dfsErr != nil {
	// 		return
	// 	}
	// 	visited.Add(subject)
	// 	if subject == object {
	// 		finalPath = *path
	// 		return
	// 	}
	// 	relations, err := u.RelationRepo.Query(domain.Relation{
	// 		SubjectNamespace: subject.Namespace,
	// 		SubjectName:      subject.Name,
	// 		SubjectRelation:  subject.Relation,
	// 	})
	// 	if err != nil {
	// 		dfsErr = err
	// 		return
	// 	}
	// 	for _, relation := range relations {
	// 		*path = append(*path, relation)
	// 		child := domain.Node{
	// 			Namespace: relation.ObjectNamespace,
	// 			Name:      relation.ObjectName,
	// 			Relation:  relation.Relation,
	// 		}
	// 		dfs(child, object, path, cur_depth+1)
	// 		*path = (*path)[:len(*path)-1]
	// 	}
	// }
	// initPath := []domain.Relation{}
	// dfs(subject, object, &initPath, 1)
	// return finalPath, nil
}

func (u *RelationUsecase) GetAllPaths(subject domain.Node, object domain.Node, searchCondition domain.SearchCondition) ([][]domain.Relation, error) {
	if err := utils.ValidateNode(object, false); err != nil {
		return nil, err
	}
	if err := utils.ValidateNode(subject, true); err != nil {
		return nil, err
	}
	paths := [][]domain.Relation{}
	type NodeItem struct {
		Cur  domain.Node
		Path []domain.Relation
	}
	firstNode := NodeItem{
		Cur:  subject,
		Path: []domain.Relation{},
	}
	q := queue.NewQueue[NodeItem]()
	q.Push(firstNode)
	for !q.IsEmpty() {
		qLen := q.Len()
		for i := 0; i < qLen; i++ {
			node, _ := q.Pop()
			query := domain.Relation{
				SubjectNamespace: node.Cur.Namespace,
				SubjectName:      node.Cur.Name,
				SubjectRelation:  node.Cur.Relation,
			}
			tuples, err := u.RelationRepo.Query(query)
			if err != nil {
				return nil, err
			}

			for _, tuple := range tuples {
				if tuple.ObjectNamespace == object.Namespace && tuple.ObjectName == object.Name && tuple.Relation == object.Relation {
					paths = append(paths, append(node.Path, tuple))
				}
				child := domain.Node{
					Namespace: tuple.ObjectNamespace,
					Name:      tuple.ObjectName,
					Relation:  tuple.Relation,
				}
				if searchCondition.ShouldStop(child) {
					continue
				}
				copyPath := append(node.Path, tuple)
				q.Push(NodeItem{
					Cur:  child,
					Path: copyPath,
				})

			}
		}
	}

	return paths, nil
}

func (u *RelationUsecase) GetAllObjectRelations(subject domain.Node, searchCondition domain.SearchCondition, collectCondition domain.CollectCondition, maxDepth int) ([]domain.Relation, error) {
	if err := utils.ValidateNode(subject, true); err != nil {
		return nil, err
	}
	depth := 0
	relations := set.NewSet[domain.Relation]()
	visited := set.NewSet[domain.Node]()
	q := queue.NewQueue[domain.Node]()
	visited.Add(subject)
	q.Push(subject)
	for !q.IsEmpty() {
		qLen := q.Len()
		for i := 0; i < qLen; i++ {
			node, _ := q.Pop()
			query := domain.Relation{
				SubjectNamespace: node.Namespace,
				SubjectName:      node.Name,
				SubjectRelation:  node.Relation,
			}
			tuples, err := u.RelationRepo.Query(query)
			if err != nil {
				return nil, err
			}

			for _, tuple := range tuples {
				child := domain.Node{
					Namespace: tuple.ObjectNamespace,
					Name:      tuple.ObjectName,
					Relation:  tuple.Relation,
				}
				if collectCondition.ShouldCollect(child) {
					relations.Add(tuple)
				}
				if !searchCondition.ShouldStop(child) && !visited.Exist(child) {
					visited.Add(child)
					q.Push(child)
				}
			}
		}
		depth++
		if depth >= maxDepth {
			break
		}
	}

	return relations.ToSlice(), nil
}

func (u *RelationUsecase) GetAllSubjectRelations(object domain.Node, searchCondition domain.SearchCondition, collectCondition domain.CollectCondition, maxDepth int) ([]domain.Relation, error) {
	if err := utils.ValidateNode(object, false); err != nil {
		return nil, err
	}
	depth := 0
	relations := set.NewSet[domain.Relation]()
	visited := set.NewSet[domain.Node]()
	q := queue.NewQueue[domain.Node]()
	visited.Add(object)
	q.Push(object)
	for !q.IsEmpty() {
		qLen := q.Len()
		for i := 0; i < qLen; i++ {
			node, _ := q.Pop()
			query := domain.Relation{
				ObjectNamespace: node.Namespace,
				ObjectName:      node.Name,
				Relation:        node.Relation,
			}
			tuples, err := u.RelationRepo.Query(query)
			if err != nil {
				return nil, err
			}

			for _, tuple := range tuples {
				parent := domain.Node{
					Namespace: tuple.SubjectNamespace,
					Name:      tuple.SubjectName,
					Relation:  tuple.SubjectRelation,
				}
				if collectCondition.ShouldCollect(parent) {
					relations.Add(tuple)
				}
				if !searchCondition.ShouldStop(parent) && !visited.Exist(parent) {
					visited.Add(parent)
					q.Push(parent)
				}
			}
		}
		depth++
		if depth >= maxDepth {
			break
		}
	}

	return relations.ToSlice(), nil
}

func (u *RelationUsecase) ClearAllRelations() error {
	return u.RelationRepo.DeleteAll()
}
