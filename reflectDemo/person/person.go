package person

type Person interface {
	Eat()
	Sleep()
	Beat()
}

//===========Person1=============

type HX struct {
	Name string
	Age int
}

func (hx HX) Eat() {
	println("hx is eating")
}

func (hx HX) Sleep() {
	println("hx is sleeping")
}

func (hx HX) Beat()  {
	println("hx: i like beat")
}

//===========Person2=============

type Wll struct {
	Name string
	Age int
}

func (wll Wll) Eat() {
	println("wll is eating")
}

func (wll Wll) Sleep() {
	println("wll is sleeping")
}

func (wll Wll) Beat()  {
	println("wll: i like beat")
}



