package service

import (
	"mooko.net/mog/model"
	"mooko.net/mog/repository"
)

/**
 * @project kuko
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-19 13:46
 * @desc
 */

var repo = repository.LibraryRepository{}

type LibraryService struct {
}

func (LibraryService) FindBy() ([]model.Library, error) {
	list, err := repo.FindBy()
	return list, err
}
