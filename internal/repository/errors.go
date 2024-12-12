package repository

import (
	"errors"
	"gorm.io/gorm"
)

var (
	Error             = errors.New("something went wrong")
	ErrRecordNotFound = errors.New("record not found")
	ErrDuplicatedKey  = errors.New("duplicated key")
)

var GormToRepositoryError = map[error]error{
	gorm.ErrRecordNotFound: ErrRecordNotFound,
	gorm.ErrDuplicatedKey:  ErrDuplicatedKey,
}

func GetRepositoryErrorByGormError(err error) error {
	if err == nil {
		return nil
	} else if GormToRepositoryError[err] != nil {
		return GormToRepositoryError[err]
	} else {
		return Error
	}
}
