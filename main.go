package main

import (
	"fmt"
)

func main() {
	// Declaración de constantes
	const pi float64 = 3.14
	const pi2 = 3.1415

	fmt.Println("pi", pi)
	fmt.Println("pi2", pi2)

	// Declaracion de variables enteras
	base := 12
	var altura int = 14
	var area int

	fmt.Println(base, altura, area)

	// Zero values
	var a int
	var b float64
	var c string
	var d bool

	fmt.Println(a, b, c, d)

	// calcular area de un cuadrado

	const baseCuadrado = 10
	areaCuadrado := baseCuadrado * baseCuadrado
	fmt.Println("Area cuadrado:", areaCuadrado)

	x := 10
	y := 50

	// suma
	result := x + y
	fmt.Println("suma:", result)

	// resta
	result = y - x
	fmt.Println("resta:", result)

	// multiplicación
	result = x * y
	fmt.Println("multiplicación:", result)

	// division
	result = y / x
	fmt.Println("division:", result)

	// modulo (es el residuo de la division)
	result = y % x
	fmt.Println("modulo:", result)

	// incrementar (+1 a la variable)
	x++
	fmt.Println("incremental:", x)

	// decremental (-1 a la variable)
	x--
	fmt.Println("decremental:", x)

	// calcular el area del resctangulo
	const largo = 10
	const ancho = 13
	areaRectangulo := largo * ancho
	fmt.Println("area rectangulo:", areaRectangulo)

	// calcular el area de un trapecio
	const base1 = 4
	const base2 = 10
	const alturaTrapecio = 7
	areaTrapecio := ((base1 + base2) / 2) * alturaTrapecio
	fmt.Println("area trapecio:", areaTrapecio)

	// calcular el area de un circulo
	const radio = 12
	areaCirculo := pi * radio
	fmt.Println("area circulo:", areaCirculo)

}

package main

func main() {
	// declaracion de variables
	helloMesage := "Hello"
	worldMesage := "World"

	// println (es un print + ln que es un saltod de linea)
	fmt.Println(helloMesage, worldMesage)

}
