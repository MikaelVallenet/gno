package home

type myInterface interface {
	Render(path string) string
	Foo() error
}

var currentImpl myInterface

func SetImpl(impl myInterface) {
	crossing()

	assertIsAdmin()
	currentImpl = impl
}

func Render(path string) string {
	assertImplIsDefined()
	return currentImpl.Render(path)
}

func Foo() error {
	assertImplIsDefined()
	return currentImpl.Foo()
}

func Bar() error {
	// doing some extra logic here
	err := currentImpl.Foo()
	// doing some more extra logic here
	return err
}

func assertImplIsDefined() {
	if currentImpl == nil {
		panic("no implementation")
	}
}

func assertIsAdmin() {
	// TODO: unsafe
}
