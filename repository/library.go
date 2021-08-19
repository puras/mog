package repository

import (
	"mooko.net/mog/model"
	"mooko.net/mog/pkg/db"
)

/**
 * @project kuko
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-19 10:15
 * @desc
 */

type LibraryRepository struct {
}

func (r *LibraryRepository) FindBy() ([]model.Library, error) {
	var ret []model.Library
	res := db.DB().Find(&ret)
	if res.Error != nil {
		return nil, res.Error
	}
	return ret, nil
}
