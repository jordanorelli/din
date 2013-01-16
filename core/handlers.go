package din

var handlerRegistry = make(map[string]Stage, 20)

func RegisterHandler(name string, stage Stage) {
	handlerRegistry[name] = stage
}

func getHandler(name string) (Stage, bool) {
	stage, ok := handlerRegistry[name]
	return stage, ok
}
