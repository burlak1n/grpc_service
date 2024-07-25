package models

type App struct {
	ID     uint32
	Name   string
	Secret string //для подписи токенов
}
