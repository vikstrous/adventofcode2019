package main

import "testing"

func BenchmarkTest(b *testing.B) {
	moons := []Moon{newMoon(
		`<x=5, y=4, z=4>`), newMoon(
		`<x=-11, y=-11, z=-3>`), newMoon(
		`<x=0, y=7, z=0>`), newMoon(
		`<x=-13, y=2, z=10>`)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		applyStep(moons, AxisX)
	}
}
