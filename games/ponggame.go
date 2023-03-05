package games

type Ball struct {
	X      int
	Y      int
	Xspeed int
	Yspeed int
}

func (b *Ball) Display() rune {
	return '\u25CF'
}

func (b *Ball) Update() {
	b.X += b.Xspeed
	b.Y += b.Yspeed
}
