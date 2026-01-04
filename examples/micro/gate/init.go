package gate

import "github.com/lonng/nano/component"

var Services = &component.Components{}

var gateService = newGateService()

func init() {
	Services.Register(gateService)
}
