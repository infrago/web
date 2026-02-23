package web

func (Router) RegistryComponent() string {
	return "web.router"
}

func (Filter) RegistryComponent() string {
	return "web.filter"
}

func (Handler) RegistryComponent() string {
	return "web.handler"
}
