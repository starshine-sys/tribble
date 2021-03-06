package etc

import (
	"image"
)

// Common colours, yoinked from discord.py
const (
	ColourTeal        = 0x1abc9c
	ColourDarkTeal    = 0x11806a
	ColourGreen       = 0x2ecc71
	ColourDarkGreen   = 0x1f8b4c
	ColourBlue        = 0x3498db
	ColourDarkBlue    = 0x206694
	ColourPurple      = 0x9b59b6
	ColourDarkPurple  = 0x71368a
	ColourMagenta     = 0xe91e63
	ColourDarkMagenta = 0xad1457
	ColourGold        = 0xf1c40f
	ColourDarkGold    = 0xc27c0e
	ColourOrange      = 0xe67e22
	ColourDarkOrange  = 0xa84300
	ColourRed         = 0xe74c3c
	ColourDarkRed     = 0x992d22
	ColourLighterGrey = 0x95a5a6
	ColourDarkGrey    = 0x607d8b
	ColourLightGrey   = 0x979c9f
	ColourDarkerGrey  = 0x546e7a
	ColourBlurple     = 0x7289da
	ColourGreyple     = 0x99aab5
	ColourDarkTheme   = 0x36393F
)

// AverageColour gets the average colour from an image.
// Return values are R, G, B, A respectively.
func AverageColour(img image.Image) (red uint8, green uint8, blue uint8, alpha uint8) {
	bounds := img.Bounds()
	minX, minY := bounds.Min.X, bounds.Min.Y
	maxX, maxY := bounds.Max.X, bounds.Max.Y

	var pixels int
	var r, g, b, a int

	for x := minX; x < maxX; x++ {
		for y := minY; y < maxY; y++ {
			if rd, gr, bl, al := img.At(x, y).RGBA(); al != 0 {
				pixels++

				r += int(rd >> 8)
				g += int(gr >> 8)
				b += int(bl >> 8)
				a += int(al >> 8)
			}
		}
	}

	if pixels == 0 {
		return 0, 0, 0, 0
	}

	return uint8(r / pixels), uint8(g / pixels), uint8(b / pixels), uint8(a / pixels)
}
