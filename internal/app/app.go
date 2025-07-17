package app

import (
	"sync"

	"exdex/config"
	"exdex/internal/router"
	"exdex/internal/src/repository"
	database "exdex/server/databases"
	"exdex/server/info"
	"exdex/server/validator"
)

type App interface {
	Start()
}

type impl struct {
	r router.Router
}

func (i *impl) Start() {

	i.Init()
	i.r.Start()

}
func NewApp(rout router.Router) App {
	return &impl{
		r: rout,
	}
}

var once sync.Once

func (i *impl) Init() {
	once.Do(func() {
		config.Init()
		validator.Init()
		database.Init()
		info.ServerInfoInit()
		repository.Init()
	})
}
