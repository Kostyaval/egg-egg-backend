package domain

type Level int

const (
	Lv0 Level = iota
	Lv1
	Lv2
	Lv3
	Lv4
	Lv5
)

func (d Level) String() string {
	return [...]string{"lv0", "lv1", "lv2", "lv3", "lv4", "lv5"}[d]
}
