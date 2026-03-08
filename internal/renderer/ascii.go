package renderer

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"

	"golang.org/x/image/draw"
)

// ─────────────────────────────────────────────
//  TEXT-TO-ASCII BLOCK FONT
//  Each character is a 5-wide × 5-tall bitmap.
//  1 = filled block (█), 0 = space.
// ─────────────────────────────────────────────

var blockFont = map[rune][5]string{
	'A': {"▐█▌", "█ █", "███", "█ █", "█ █"},
	'B': {"██▌", "█ █", "███", "█ █", "██▌"},
	'C': {"▐██", "█  ", "█  ", "█  ", "▐██"},
	'D': {"██▌", "█ █", "█ █", "█ █", "██▌"},
	'E': {"███", "█  ", "██▌", "█  ", "███"},
	'F': {"███", "█  ", "██▌", "█  ", "█  "},
	'G': {"▐██", "█  ", "█▐█", "█ █", "▐██"},
	'H': {"█ █", "█ █", "███", "█ █", "█ █"},
	'I': {"███", " █ ", " █ ", " █ ", "███"},
	'J': {"███", "  █", "  █", "█ █", "▐█▌"},
	'K': {"█ █", "█▐▌", "██ ", "█▐▌", "█ █"},
	'L': {"█  ", "█  ", "█  ", "█  ", "███"},
	'M': {"█ █", "███", "█▐█", "█ █", "█ █"},
	'N': {"█ █", "██▌", "█▐█", "█ █", "█ █"},
	'O': {"▐█▌", "█ █", "█ █", "█ █", "▐█▌"},
	'P': {"██▌", "█ █", "███", "█  ", "█  "},
	'Q': {"▐█▌", "█ █", "█ █", "█▐▌", "▐██"},
	'R': {"██▌", "█ █", "███", "██ ", "█ █"},
	'S': {"▐██", "█  ", "▐█▌", "  █", "██▌"},
	'T': {"███", " █ ", " █ ", " █ ", " █ "},
	'U': {"█ █", "█ █", "█ █", "█ █", "▐█▌"},
	'V': {"█ █", "█ █", "█ █", "▐█▌", " █ "},
	'W': {"█ █", "█ █", "█▐█", "███", "█ █"},
	'X': {"█ █", "▐█▌", " █ ", "▐█▌", "█ █"},
	'Y': {"█ █", "▐█▌", " █ ", " █ ", " █ "},
	'Z': {"███", "  █", " █ ", "█  ", "███"},
	'0': {"▐█▌", "█ █", "█ █", "█ █", "▐█▌"},
	'1': {" █ ", "██ ", " █ ", " █ ", "███"},
	'2': {"▐█▌", "█ █", " ▐█", "█▌ ", "███"},
	'3': {"██▌", "  █", " █▌", "  █", "██▌"},
	'4': {"█ █", "█ █", "███", "  █", "  █"},
	'5': {"███", "█  ", "██▌", "  █", "██▌"},
	'6': {"▐█▌", "█  ", "██▌", "█ █", "▐█▌"},
	'7': {"███", "  █", " █ ", " █ ", " █ "},
	'8': {"▐█▌", "█ █", "▐█▌", "█ █", "▐█▌"},
	'9': {"▐█▌", "█ █", "▐██", "  █", "▐█▌"},
	' ': {"   ", "   ", "   ", "   ", "   "},
	'.': {"   ", "   ", "   ", "   ", " █ "},
	'!': {" █ ", " █ ", " █ ", "   ", " █ "},
	'-': {"   ", "   ", "███", "   ", "   "},
	'_': {"   ", "   ", "   ", "   ", "███"},
	':': {"   ", " █ ", "   ", " █ ", "   "},
	'/': {"  █", " █ ", " █ ", "█  ", "█  "},
}

// TextToASCII renders text as large ANSI block characters.
// Returns a multiline string with colour (bold cyan).
func TextToASCII(text string) string {
	text = strings.ToUpper(text)
	rows := [5]strings.Builder{}

	for i, ch := range text {
		glyph, ok := blockFont[ch]
		if !ok {
			glyph = blockFont[' ']
		}
		for r := 0; r < 5; r++ {
			rows[r].WriteString(glyph[r])
			if i < len([]rune(text))-1 {
				rows[r].WriteString(" ")
			}
		}
	}

	const cyan = "\033[1;96m"
	const reset = "\033[0m"
	var sb strings.Builder
	for _, r := range rows {
		sb.WriteString(cyan + r.String() + reset + "\n")
	}
	return sb.String()
}

// ─────────────────────────────────────────────
//  IMAGE TO HALF-BLOCK RENDERER
//  Uses the ▀ character:
//    foreground = top pixel colour
//    background = bottom pixel colour
//  This packs 2 vertical pixels per terminal row.
// ─────────────────────────────────────────────

// ImageToHalfBlock reads imagePath, resizes to `width` characters wide,
// and returns an ANSI true-colour string using half-block art.
func ImageToHalfBlock(imagePath string, width int) (string, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("open image: %w", err)
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}

	// Height is halved (two pixels per row) and we keep ~1:2 aspect ratio.
	height := (width * src.Bounds().Dy() / src.Bounds().Dx()) & ^1 // ensure even

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	const reset = "\033[0m"
	var sb strings.Builder

	for y := 0; y < height; y += 2 {
		for x := 0; x < width; x++ {
			top := dst.RGBAAt(x, y)
			bot := dst.RGBAAt(x, y+1)
			// fg = top pixel, bg = bottom pixel
			fmt.Fprintf(&sb,
				"\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm▀",
				top.R, top.G, top.B,
				bot.R, bot.G, bot.B,
			)
		}
		sb.WriteString(reset + "\n")
	}

	return sb.String(), nil
}
