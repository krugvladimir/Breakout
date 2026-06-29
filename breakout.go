// breakout.go
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
	"github.com/nsf/termbox-go"
)

const recordFile = ".breakout_record.json"
type Record struct{ Best int }

func loadRecord() int {
	f, err := os.Open(recordFile)
	if err != nil { return 0 }
	defer f.Close()
	var r Record
	json.NewDecoder(f).Decode(&r)
	return r.Best
}
func saveRecord(best int) {
	f, _ := os.Create(recordFile)
	defer f.Close()
	json.NewEncoder(f).Encode(Record{best})
}

type Block struct{ x, y, hp, color int }
type Bonus struct{ x, y, typ int }

func main() {
	speed := 50
	if len(os.Args) > 2 && os.Args[1] == "-s" {
		if s, err := strconv.Atoi(os.Args[2]); err == nil && s > 0 { speed = s }
	}
	err := termbox.Init()
	if err != nil { fmt.Println(err); return }
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)
	w, h := termbox.Size()
	if h < 25 || w < 50 { fmt.Println("Terminal too small"); return }
	rand.Seed(time.Now().UnixNano())

	H, W := h-2, w-2
	padY, padX := 1, 1
	paddleW := 8
	paddleX := (W - paddleW) / 2
	paddleY := H - 2
	ballX := float64(paddleX + paddleW/2)
	ballY := float64(paddleY - 1)
	ballDX, ballDY := 1.0, -1.0
	ballSpeed := 1.5
	ballActive := false
	gameOver := false
	lives, score, level := 3, 0, 1
	best := loadRecord()
	frame := time.Duration(speed) * time.Millisecond

	blocks := []Block{}
	bonuses := []Bonus{}

	generateLevel := func() {
		blocks = []Block{}
		rows := 5 + level/2
		cols := 8 + level
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				if c*2+2 < W-2 {
					hp := 1
					if rand.Intn(100) < 30 {
						hp = 2
					} else if rand.Intn(100) < 10 {
						hp = 3
					}
					color := hp + 1
					if color > 7 {
						color = 7
					}
					blocks = append(blocks, Block{c*2 + 2, r*2 + 2, hp, color})
				}
			}
		}
	}
	generateLevel()

	tbprint := func(x, y int, fg, bg termbox.Attribute, msg string) {
		for _, ch := range msg {
			termbox.SetCell(x, y, ch, fg, bg)
			x++
		}
	}

	for {
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			if ev.Key == termbox.KeyEsc || ev.Ch == 'q' { return }
			if ev.Ch == 'r' && gameOver {
				paddleX = (W - paddleW) / 2
				ballX = float64(paddleX + paddleW/2)
				ballY = float64(paddleY - 1)
				ballDX, ballDY = 1.0, -1.0
				ballActive = false
				lives, score, level = 3, 0, 1
				gameOver = false
				bonuses = []Bonus{}
				paddleW = 8
				ballSpeed = 1.5
				generateLevel()
				continue
			}
			if ev.Ch == ' ' {
				if !ballActive && !gameOver {
					ballActive = true
				}
			}
			if ev.Key == termbox.KeyArrowLeft || ev.Ch == 'a' {
				paddleX = max(0, paddleX-2)
			}
			if ev.Key == termbox.KeyArrowRight || ev.Ch == 'd' {
				paddleX = min(W-paddleW, paddleX+2)
			}
		}

		if gameOver {
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			msg := fmt.Sprintf("GAME OVER! Score: %d  Best: %d", score, best)
			tbprint((w-len(msg))/2, h/2-2, termbox.ColorWhite, termbox.ColorDefault, msg)
			tbprint((w-20)/2, h/2, termbox.ColorCyan, termbox.ColorDefault, "R - restart | Q - quit")
			termbox.Flush()
			continue
		}

		if ballActive {
			ballX += ballDX * ballSpeed
			ballY += ballDY * ballSpeed
			if ballX <= 0 || ballX >= float64(W-1) {
				ballDX *= -1
			}
			if ballY <= 0 {
				ballDY *= -1
			}
			if int(ballY) == paddleY-1 && paddleX <= int(ballX) && int(ballX) < paddleX+paddleW {
				ballDY *= -1
				offset := (ballX - float64(paddleX)) / float64(paddleW)
				ballDX = (offset - 0.5) * 2
				if absFloat(ballDX) < 0.3 {
					ballDX = 0.5
					if ballDX < 0 {
						ballDX = -0.5
					}
				}
				fmt.Print("\a")
			}
			for i := 0; i < len(blocks); {
				b := &blocks[i]
				if float64(b.x) <= ballX && ballX < float64(b.x+2) &&
					float64(b.y) <= ballY && ballY < float64(b.y+2) {
					b.hp--
					if b.hp <= 0 {
						score += 10
						if rand.Intn(100) < 15 {
							btype := rand.Intn(4)
							bonuses = append(bonuses, Bonus{b.x, b.y, btype})
						}
						blocks = append(blocks[:i], blocks[i+1:]...)
						fmt.Print("\a")
					} else {
						blocks[i] = *b
						i++
					}
					ballDY *= -1
					break
				} else {
					i++
				}
			}
			if ballY >= float64(H) {
				lives--
				if lives <= 0 {
					gameOver = true
					if score > best {
						best = score
						saveRecord(best)
					}
				} else {
					ballX = float64(paddleX + paddleW/2)
					ballY = float64(paddleY - 1)
					ballDX, ballDY = 1.0, -1.0
					ballActive = false
				}
			}
			if len(blocks) == 0 {
				level++
				ballSpeed += 0.3
				generateLevel()
				ballX = float64(paddleX + paddleW/2)
				ballY = float64(paddleY - 1)
				ballDX, ballDY = 1.0, -1.0
				ballActive = false
			}
		}

		for i := 0; i < len(bonuses); i++ {
			b := &bonuses[i]
			b.y++
			if b.y >= H {
				bonuses = append(bonuses[:i], bonuses[i+1:]...)
				i--
				continue
			}
			if b.y == paddleY && paddleX <= b.x && b.x < paddleX+paddleW {
				if b.typ == 0 {
					paddleW = min(16, paddleW+4)
				} else if b.typ == 1 {
					ballSpeed += 0.5
				} else if b.typ == 2 {
					lives++
				} else if b.typ == 3 {
					for j := range blocks {
						blocks[j].hp--
					}
				}
				bonuses = append(bonuses[:i], bonuses[i+1:]...)
				i--
				fmt.Print("\a")
			}
		}

		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		// Рамка
		for y := 0; y <= H; y++ {
			termbox.SetCell(padX-1, padY+y, '|', termbox.ColorWhite, termbox.ColorDefault)
			termbox.SetCell(padX+W, padY+y, '|', termbox.ColorWhite, termbox.ColorDefault)
		}
		for x := 0; x < W+2; x++ {
			termbox.SetCell(padX+x-1, padY-1, '-', termbox.ColorWhite, termbox.ColorDefault)
		}
		// Платформа
		for i := 0; i < paddleW; i++ {
			termbox.SetCell(padX+paddleX+i, padY+paddleY, '=', termbox.ColorYellow, termbox.ColorDefault)
		}
		// Мяч
		if ballActive || !gameOver {
			termbox.SetCell(padX+int(ballX), padY+int(ballY), 'O', termbox.ColorCyan, termbox.ColorDefault)
		}
		// Блоки
		for _, b := range blocks {
			for i := 0; i < 2; i++ {
				for j := 0; j < 2; j++ {
					termbox.SetCell(padX+b.x+i, padY+b.y+j, '#', termbox.Attribute(b.color), termbox.ColorDefault)
				}
			}
		}
		// Бонусы
		for _, bon := range bonuses {
			sym := 'W'
			if bon.typ == 1 {
				sym = 'S'
			} else if bon.typ == 2 {
				sym = 'L'
			} else if bon.typ == 3 {
				sym = 'F'
			}
			termbox.SetCell(padX+bon.x, padY+bon.y, sym, termbox.ColorMagenta, termbox.ColorDefault)
		}
		tbprint(2, 0, termbox.ColorWhite, termbox.ColorDefault, fmt.Sprintf("Score: %d", score))
		tbprint(W/2-4, 0, termbox.ColorWhite, termbox.ColorDefault, fmt.Sprintf("Best: %d", best))
		tbprint(W-20, 0, termbox.ColorWhite, termbox.ColorDefault, fmt.Sprintf("Lives: %d  Level: %d", lives, level))
		if gameOver {
			msg := "GAME OVER! Press R to restart, Q to quit"
			tbprint((w-len(msg))/2, h/2, termbox.ColorRed, termbox.ColorDefault, msg)
		}
		termbox.Flush()
		time.Sleep(frame)
	}
}
func max(a, b int) int { if a > b { return a }; return b }
func min(a, b int) int { if a < b { return a }; return b }
func absFloat(a float64) float64 { if a < 0 { return -a }; return a }
