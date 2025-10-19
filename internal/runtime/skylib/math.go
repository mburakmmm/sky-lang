package skylib

/*
#include <math.h>
*/
import "C"
import "math"

// Mathematical constants
const (
	PI  = math.Pi
	E   = math.E
	TAU = 2 * math.Pi
)

// Basic math functions
func Abs(x float64) float64 {
	return math.Abs(x)
}

func Min(a, b float64) float64 {
	return math.Min(a, b)
}

func Max(a, b float64) float64 {
	return math.Max(a, b)
}

func Pow(x, y float64) float64 {
	return math.Pow(x, y)
}

func Sqrt(x float64) float64 {
	return math.Sqrt(x)
}

func Floor(x float64) float64 {
	return math.Floor(x)
}

func Ceil(x float64) float64 {
	return math.Ceil(x)
}

func Round(x float64) float64 {
	return math.Round(x)
}

// Trigonometric functions
func Sin(x float64) float64 {
	return math.Sin(x)
}

func Cos(x float64) float64 {
	return math.Cos(x)
}

func Tan(x float64) float64 {
	return math.Tan(x)
}

func Asin(x float64) float64 {
	return math.Asin(x)
}

func Acos(x float64) float64 {
	return math.Acos(x)
}

func Atan(x float64) float64 {
	return math.Atan(x)
}

func Atan2(y, x float64) float64 {
	return math.Atan2(y, x)
}

// Exponential and logarithmic functions
func Exp(x float64) float64 {
	return math.Exp(x)
}

func Log(x float64) float64 {
	return math.Log(x)
}

func Log10(x float64) float64 {
	return math.Log10(x)
}

func Log2(x float64) float64 {
	return math.Log2(x)
}

// Hyperbolic functions
func Sinh(x float64) float64 {
	return math.Sinh(x)
}

func Cosh(x float64) float64 {
	return math.Cosh(x)
}

func Tanh(x float64) float64 {
	return math.Tanh(x)
}

// Utility functions
func Hypot(x, y float64) float64 {
	return math.Hypot(x, y)
}

func Mod(x, y float64) float64 {
	return math.Mod(x, y)
}

func IsNaN(x float64) bool {
	return math.IsNaN(x)
}

func IsInf(x float64) bool {
	return math.IsInf(x, 0)
}
