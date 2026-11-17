package repository

import "gorm.io/gorm"

type crudRepo[T any] interface {
	Create(v *T) error
	Update(v *T) error
	Delete(id int) error
	FindByID(id int) (*T, error)
	List() ([]T, error)
}

type crudRepository[T any] struct {
	db    *gorm.DB
	order string
}

func (r *crudRepository[T]) Create(v *T) error { return r.db.Create(v).Error }
func (r *crudRepository[T]) Update(v *T) error { return r.db.Save(v).Error }
func (r *crudRepository[T]) Delete(id int) error {
	var zero T
	return r.db.Delete(&zero, id).Error
}

func (r *crudRepository[T]) FindByID(id int) (*T, error) {
	var v T
	if err := r.db.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *crudRepository[T]) List() ([]T, error) {
	var vs []T
	err := r.db.Order(r.order).Find(&vs).Error
	return vs, err
}
