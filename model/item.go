package model

type Item struct {
	Name string
}

func NewItem() Item {
	return Item{Name: ""}
}
