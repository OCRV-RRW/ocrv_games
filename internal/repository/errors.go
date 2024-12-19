package repository

import (
	"errors"
	"gorm.io/gorm"
	"strings"
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
		if strings.Contains(err.Error(), "duplicate key value violates unique") {
			return ErrDuplicatedKey
		}
		return Error
	}
}
