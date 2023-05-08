package main

import (
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Meal struct {
	Name         string
	ImageUrl     string
	Ingredients  string
	Calories     int
	Protein      int
	Fat          int
	SaturatedFat int
	Carbs        int
	Sugar        int
	Cholestrol   int
	Sodium       int
}

func main() {
	page := rod.New().MustConnect().MustPage("https://app.yumba.ca/#/on-the-menu")
	page.MustWaitLoad()
	mealCards, err := page.Elements(".meal-card")
	if err != nil {
		panic(1)
	}
	for _, mealCard := range mealCards {
		span, err := mealCard.Element("span")
		if err != nil {
			panic(1)
		}
		if err := span.Click(proto.InputMouseButtonLeft, 1); err != nil {
			panic(1)
		}
		modal, err := page.Element(".meal-modal")
		for err != nil {
			modal, err = page.Element(".meal-modal")
		}
		fmt.Println(modal.Text())
		if err != nil {
			panic(1)
		}
		fmt.Println(modal.Text())
		closeBtn, err := modal.Element(".close-modal")
		if err != nil {
			panic(1)
		}
		if err := closeBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
			panic(1)
		}
	}
}
