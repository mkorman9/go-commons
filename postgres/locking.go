package postgres

import "gorm.io/gorm"

type LockingService interface {
	Acquire(lockID int64) (bool, error)
	Release(lockID int64) error
}

type lockingService struct {
	db *gorm.DB
}

func NewLockingService(db *gorm.DB) LockingService {
	return &lockingService{db}
}

func (service *lockingService) Acquire(lockID int64) (bool, error) {
	ok := false
	if err := service.db.Raw("SELECT pg_try_advisory_lock(?)", lockID).Scan(&ok).Error; err != nil {
		return false, err
	}

	return ok, nil
}

func (service *lockingService) Release(lockID int64) error {
	if err := service.db.Exec("SELECT pg_advisory_unlock(?)", lockID).Error; err != nil {
		return err
	}

	return nil
}
