package mpeg7cld

import (
	"image"
	"image/color"
	"math"
)

// Colour Layout Descriptor
// see: https://en.wikipedia.org/wiki/Color_layout_descriptor
func CLD(img image.Image) [64]YCbCr {
	partitions := partition(img)
	averages := average(partitions)
	ycbcr := rgb2ycbcr(averages)
	dct := dct(ycbcr)
	return zigzag(dct)
}

type rgb struct {
	r, g, b uint8
}

func (c rgb) RGBA() (r, g, b, a uint32) {
	return uint32(c.r), uint32(c.g), uint32(c.b), 0xFF
}

type YCbCr struct {
	Y, Cb, Cr float64
}

func partition(img image.Image) [64][]color.Color {
	var blocks [64][]color.Color

	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y

	min := func(a, b int) int {
		if a <= b {
			return a
		}
		return b
	}

	partitionWidth := int(float64(width) / 8)
	partitionHeight := int(float64(height) / 8)
	var x, y, i int
	for x = 0; x < partitionWidth*8; x += partitionWidth {
		for y = 0; y < partitionHeight*8; y += partitionHeight {
			w := min(partitionWidth, width-x)
			h := min(partitionHeight, height-y)
			blocks[i] = make([]color.Color, 0, w*h)
			for dx := x; dx < x+w; dx++ {
				for dy := y; dy < y+h; dy++ {
					blocks[i] = append(blocks[i], img.At(dx, dy))
				}
			}
			i++
		}
	}
	return blocks
}

func average(partitions [64][]color.Color) [64]rgb {
	var blocks [64]rgb
	for i, partition := range partitions {
		var sumRed, sumBlue, sumGreen uint32
		for _, c := range partition {
			r, g, b, _ := c.RGBA()
			sumRed += r >> 8
			sumBlue += b >> 8
			sumGreen += g >> 8
		}
		blocks[i] = rgb{
			r: uint8(int(sumRed) / len(partition)),
			g: uint8(int(sumBlue) / len(partition)),
			b: uint8(int(sumGreen) / len(partition)),
		}
	}
	return blocks
}

func rgb2ycbcr(img [64]rgb) [64]YCbCr {
	var blocks [64]YCbCr
	for i, p := range img {
		y, cb, cr := color.RGBToYCbCr(p.r, p.g, p.b)
		blocks[i] = YCbCr{
			Y:  float64(y),
			Cb: float64(cb),
			Cr: float64(cr),
		}
	}
	return blocks
}

func dct(in [64]YCbCr) [64]YCbCr {
	var out [64]YCbCr
	for p := 0; p < 8; p++ {
		for q := 0; q < 8; q++ {
			var alphaP float64
			if p > 0 {
				alphaP = math.Sqrt(2.0 / 8.0)
			} else {
				alphaP = float64(1.0) / math.Sqrt(8)
			}

			var alphaQ float64
			if q > 0 {
				alphaQ = math.Sqrt(2.0 / 8.0)
			} else {
				alphaQ = float64(1.0) / math.Sqrt(8)
			}

			var sumY, sumCb, sumCr float64
			for m := 0; m < 8; m++ {
				for n := 0; n < 8; n++ {
					c := math.Cos(math.Pi*(2*float64(m)+1)*float64(p)/16.0) * math.Cos(math.Pi*(2*float64(n)+1)*float64(q)/16.0)
					i := index(m, n)
					sumY += float64(in[i].Y) * c
					sumCb += float64(in[i].Cb) * c
					sumCr += float64(in[i].Cr) * c
				}
			}
			i := index(p, q)
			out[i].Y = alphaP * alphaQ * sumY
			out[i].Cb = alphaP * alphaQ * sumCb
			out[i].Cr = alphaP * alphaQ * sumCr
		}
	}
	return out
}

type direction int

const (
	DOWN direction = iota
	UP
)

func zigzag(in [64]YCbCr) [64]YCbCr {
	var (
		x, y, i   int
		direction direction = UP
	)

	for {
		j := index(y, x)
		in[i], in[j] = in[j], in[i]
		i++

		if x == 7 && y == 7 {
			break
		}

		switch direction {
		case UP:
			switch {
			case x == 7:
				y++
				direction = DOWN
			case y == 0:
				x++
				direction = DOWN
			default:
				x++
				y--
			}
		case DOWN:
			switch {
			case y == 7:
				x++
				direction = UP
			case x == 0:
				y++
				direction = UP
			default:
				x--
				y++
			}
		}
	}
	return in
}

func index(x, y int) int {
	i := y + 8*x
	return int(i)
}

func Compare(cld1, cld2 [64]YCbCr) float64 {
	var result float64
	for i := 0; i < 64; i++ {
		result += math.Sqrt((cld1[i].Y - cld2[i].Y) * (cld1[i].Y - cld2[i].Y))
		result += math.Sqrt((cld1[i].Cb - cld2[i].Cb) * (cld1[i].Cb - cld2[i].Cb))
		result += math.Sqrt((cld1[i].Cr - cld2[i].Cr) * (cld1[i].Cr - cld2[i].Cr))
	}
	return result
}
