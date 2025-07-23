package handler

import "exdex/internal/src/services"

func WsInit() {
	services.OrderWSHandler()
}
