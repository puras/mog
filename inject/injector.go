package inject

import (
	"github.com/google/wire"
	"github.com/puras/mog/dbx"
	"github.com/puras/mog/jwtx"
	"gorm.io/gorm"
)

type Injector struct {
	DB   *gorm.DB
	Auth jwtx.Auther
}

func InitSet() wire.ProviderSet {
	return wire.NewSet(
		dbx.InitDB,
		jwtx.InitAuth,
		wire.NewSet(wire.Struct(new(Injector), "*")),
		wire.NewSet(wire.Struct(new(dbx.Trans), "*")),
	)
}
