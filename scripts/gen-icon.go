//go:build ignore

package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

func main() {
	for _, dir := range []string{
		"build",
		filepath.Join("build", "windows"),
		filepath.Join("frontend", "src", "assets", "images"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			panic(err)
		}
	}

	img512 := renderIcon(512)

	if err := savePNG(filepath.Join("build", "appicon.png"), img512); err != nil {
		panic(err)
	}
	if err := savePNG(filepath.Join("frontend", "src", "assets", "images", "logo-universal.png"), img512); err != nil {
		panic(err)
	}

	sizes := []int{256, 64, 48, 32, 16}
	if err := saveICO(filepath.Join("build", "windows", "icon.ico"), sizes); err != nil {
		panic(err)
	}
}

func renderIcon(size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	s := float64(size)
	cx, cy := s/2.0, s/2.0
	margin := s * 0.055
	halfBox := s/2.0 - margin
	cornerR := s * 0.22

	// Play triangle vertices centered slightly right for optical balance
	tx0, ty0 := cx-s*0.11, cy-s*0.16
	tx1, ty1 := cx-s*0.11, cy+s*0.16
	tx2, ty2 := cx+s*0.17, cy

	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			x := float64(px) + 0.5
			y := float64(py) + 0.5

			// Signed distance to rounded squircle
			dx := math.Abs(x-cx) - (halfBox - cornerR)
			dy := math.Abs(y-cy) - (halfBox - cornerR)
			qx := math.Max(dx, 0)
			qy := math.Max(dy, 0)
			sdf := math.Hypot(qx, qy) + math.Min(math.Max(dx, dy), 0) - cornerR

			if sdf > 1.0 {
				continue
			}

			alpha := uint8(255)
			if sdf > -1.0 {
				a := 1.0 - (sdf+1.0)/2.0
				if a < 0 {
					a = 0
				}
				if a > 1 {
					a = 1
				}
				alpha = uint8(a * 255)
			}

			// Base dark background #090D16 with subtle violet/cyan radial glow
			dist := math.Hypot(x-cx, y-cy) / (s * 0.5)
			glow := 1.0 - math.Min(dist, 1.0)
			r := 9.0 + glow*22.0
			g := 13.0 + glow*14.0
			b := 22.0 + glow*44.0

			// Outer squircle border highlight (Violet #8B5CF6)
			if sdf > -s*0.022 {
				r = mix(r, 139, 0.35)
				g = mix(g, 92, 0.35)
				b = mix(b, 246, 0.35)
			}

			// Outer luminous ring (Violet -> Cyan gradient along angle)
			rad := math.Hypot(x-cx, y-cy)
			ringW := math.Max(2.0, s*0.024)
			ang := math.Atan2(y-cy, x-cx)
			gradT := (math.Cos(ang) + 1.0) * 0.5 // 0..1
			ringRColor := mix(139, 6, gradT)
			ringGColor := mix(92, 182, gradT)
			ringBColor := mix(246, 212, gradT)

			dOuter := math.Abs(rad - s*0.35)
			if dOuter < ringW*1.8 {
				t := math.Exp(-(dOuter * dOuter) / (ringW * ringW * 0.55))
				r = mix(r, ringRColor, t*0.95)
				g = mix(g, ringGColor, t*0.95)
				b = mix(b, ringBColor, t*0.95)
			}

			// Subtle inner ring
			dInner := math.Abs(rad - s*0.26)
			if dInner < ringW*1.2 {
				t := math.Exp(-(dInner * dInner) / (ringW * ringW * 0.45))
				r = mix(r, 6, t*0.35)
				g = mix(g, 182, t*0.35)
				b = mix(b, 212, t*0.35)
			}

			// Play triangle with soft rounded glow
			dTri := sdTriangle(x, y, tx0, ty0, tx1, ty1, tx2, ty2)
			triCorner := s * 0.022
			dTriRound := dTri - triCorner
			if dTriRound < s*0.04 {
				if dTriRound <= 0 {
					// Inside play button: bright cyan-to-violet vertical gradient
					vt := clamp01((y - (cy - s*0.16)) / (s * 0.32))
					tr := mix(34, 167, vt)
					tg := mix(211, 139, vt)
					tb := mix(238, 250, vt)
					edge := clamp01(-dTriRound / math.Max(1.0, s*0.006))
					r = mix(r, tr, edge)
					g = mix(g, tg, edge)
					b = mix(b, tb, edge)
				} else {
					glowT := math.Exp(-(dTriRound * dTriRound) / (s * s * 0.0005))
					r = mix(r, 56, glowT*0.45)
					g = mix(g, 189, glowT*0.45)
					b = mix(b, 248, glowT*0.45)
				}
			}

			// Subtle pulse wave across bottom-center inside ring
			nx := (x - cx) / (s * 0.30)
			if math.Abs(nx) <= 0.95 && y > cy+s*0.18 && y < cy+s*0.31 {
				pyNorm := cy + s*0.24 - math.Sin(nx*math.Pi*2.0)*math.Exp(-nx*nx*2.5)*s*0.04
				dWave := math.Abs(y - pyNorm)
				strokeW := math.Max(1.5, s*0.015)
				if dWave < strokeW*1.6 {
					t := math.Exp(-(dWave * dWave) / (strokeW * strokeW * 0.5))
					r = mix(r, 16, t*0.9)
					g = mix(g, 185, t*0.9)
					b = mix(b, 129, t*0.9)
				}
			}

			img.SetNRGBA(px, py, color.NRGBA{
				R: clampU8(r),
				G: clampU8(g),
				B: clampU8(b),
				A: alpha,
			})
		}
	}
	return img
}

func sdTriangle(px, py, ax, ay, bx, by, cx, cy float64) float64 {
	e0x, e0y := bx-ax, by-ay
	e1x, e1y := cx-bx, cy-by
	e2x, e2y := ax-cx, ay-cy

	v0x, v0y := px-ax, py-ay
	v1x, v1y := px-bx, py-by
	v2x, v2y := px-cx, py-cy

	pq0x, pq0y := v0x-e0x*clamp01(dot(v0x, v0y, e0x, e0y)/dot(e0x, e0y, e0x, e0y)), v0y-e0y*clamp01(dot(v0x, v0y, e0x, e0y)/dot(e0x, e0y, e0x, e0y))
	pq1x, pq1y := v1x-e1x*clamp01(dot(v1x, v1y, e1x, e1y)/dot(e1x, e1y, e1x, e1y)), v1y-e1y*clamp01(dot(v1x, v1y, e1x, e1y)/dot(e1x, e1y, e1x, e1y))
	pq2x, pq2y := v2x-e2x*clamp01(dot(v2x, v2y, e2x, e2y)/dot(e2x, e2y, e2x, e2y)), v2y-e2y*clamp01(dot(v2x, v2y, e2x, e2y)/dot(e2x, e2y, e2x, e2y))

	s := 1.0
	if e0x*e2y-e0y*e2x < 0 {
		s = -1.0
	}

	d0 := dot(pq0x, pq0y, pq0x, pq0y)
	d1 := dot(pq1x, pq1y, pq1x, pq1y)
	d2 := dot(pq2x, pq2y, pq2x, pq2y)

	s0 := s * (v0x*e0y - v0y*e0x)
	s1 := s * (v1x*e1y - v1y*e1x)
	s2 := s * (v2x*e2y - v2y*e2x)

	minD := math.Min(d0, math.Min(d1, d2))
	minS := math.Min(s0, math.Min(s1, s2))
	sign := 1.0
	if minS > 0 {
		sign = -1.0
	}
	return sign * math.Sqrt(minD)
}

func dot(ax, ay, bx, by float64) float64 {
	return ax*bx + ay*by
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func mix(a, b, t float64) float64 {
	return a + (b-a)*t
}

func clampU8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func saveICO(path string, sizes []int) error {
	type icoFrame struct {
		size int
		png  []byte
	}
	frames := make([]icoFrame, 0, len(sizes))
	for _, sz := range sizes {
		var buf bytes.Buffer
		if err := png.Encode(&buf, renderIcon(sz)); err != nil {
			return err
		}
		frames = append(frames, icoFrame{size: sz, png: buf.Bytes()})
	}

	var out bytes.Buffer
	_ = binary.Write(&out, binary.LittleEndian, uint16(0))
	_ = binary.Write(&out, binary.LittleEndian, uint16(1))
	_ = binary.Write(&out, binary.LittleEndian, uint16(len(frames)))

	offset := uint32(6 + 16*len(frames))
	for _, fr := range frames {
		dim := byte(fr.size)
		if fr.size >= 256 {
			dim = 0
		}
		out.WriteByte(dim)
		out.WriteByte(dim)
		out.WriteByte(0)
		out.WriteByte(0)
		_ = binary.Write(&out, binary.LittleEndian, uint16(1))
		_ = binary.Write(&out, binary.LittleEndian, uint16(32))
		_ = binary.Write(&out, binary.LittleEndian, uint32(len(fr.png)))
		_ = binary.Write(&out, binary.LittleEndian, offset)
		offset += uint32(len(fr.png))
	}
	for _, fr := range frames {
		out.Write(fr.png)
	}

	return os.WriteFile(path, out.Bytes(), 0o644)
}
