package main

import "testing"

func TestObtenerRetoBinding(t *testing.T) {
	ch := obtenerRetoBinding()

	if ch.Question == "" {
		t.Fatal("field Question is empty")
	}
	if ch.Answer == "" {
		t.Fatal("field Answer is empty")
	}
	if ch.Key == "" {
		t.Fatal("field Key is empty")
	}
}
