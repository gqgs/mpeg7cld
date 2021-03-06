package mpeg7cld

import (
	"image/jpeg"
	"os"
	"testing"
)

func Test_Compare(t *testing.T) {
	file, err := os.Open("testdata/jpeg_image.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		t.Fatal(err)
	}

	cld := CLD(img)
	if res := Compare(cld, cld); res != 0 {
		t.Errorf("Compare(): want: 0; got: %0.2f", res)
	}
}

func Benchmark_partition(b *testing.B) {
	file, err := os.Open("testdata/jpeg_image.jpg")
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = partition(img)
	}
}
