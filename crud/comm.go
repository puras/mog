package crud

import "github.com/puras/mog/dbx"

type QueryParams interface {
	GetPaginationParam() dbx.PaginationParam
	//SetPaginationParam(*dbx.PaginationParam)
}

type PageParams struct {
	dbx.PaginationParam
}

func (self PageParams) GetPaginationParam() dbx.PaginationParam {
	return self.PaginationParam
}
