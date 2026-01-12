package main

type Vector2 struct {
	X, Y float32
}

func lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}
