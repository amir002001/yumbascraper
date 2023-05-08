package main

import (
	"fmt"

	"github.com/go-rod/rod"
)

func main() {
	page := rod.New().MustConnect().MustPage("https://app.yumba.ca/#/on-the-menu")
	page.MustWaitLoad()
	mealCards, err := page.Elements(".meal-card")
	if err != nil {
		panic(1)
	}
	for _, mealCard := range mealCards {
		fmt.Println(mealCard.Text())
	}
}
