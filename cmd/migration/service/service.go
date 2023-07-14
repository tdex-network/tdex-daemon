package service

type Service interface {
	Migrate() error
}
