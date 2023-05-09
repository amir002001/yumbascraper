package main

import (
	"fmt"
	"strconv"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Meal struct {
	Name        string
	ImageUrl    string
	Ingredients string
	Calories    int
	Protein     int
	Carbs       int
}

func main() {
	page := rod.New().MustConnect().MustPage("https://app.yumba.ca/#/on-the-menu")
	page.MustWaitLoad()
	mealCards, err := page.Elements(".meal-card")
	if err != nil {
		panic(err)
	}
	meals := []Meal{}
	for _, mealCard := range mealCards {
		span, err := mealCard.Element("span")
		if err != nil {
			panic(err)
		}
		if err := span.Click(proto.InputMouseButtonLeft, 1); err != nil {
			panic(err)
		}
		modal, err := page.Element(".meal-modal")
		for err != nil {
			modal, err = page.Element(".meal-modal")
		}
		if err != nil {
			panic(err)
		}
		meal, err := createMealFromModal(modal)
		if err != nil {
			panic(err)
		}
		meals = append(meals, meal)
		// close the modal
		closeBtn, err := modal.Element(".close-modal")
		if err != nil {
			panic(err)
		}
		if err := closeBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
			panic(err)
		}
	}
	for i := 0; i < len(meals); i++ {
		err := addMealToNotion(&meals[i])
		if err != nil {
			panic(err)
		}
	}
}

func createMealFromModal(modal *rod.Element) (Meal, error) {
	var meal Meal
	calories, carbs, protein, err := extractStatsFromModal(modal)
	if err != nil {
		return meal, err
	}
	ingredients, err := extractIngredientsFromModal(modal)
	if err != nil {
		return meal, err
	}
	name, err := extractNameFromModal(modal)
	if err != nil {
		return meal, err
	}
	src, err := extractImageFromModal(modal)
	if err != nil {
		return meal, err
	}
	meal.Name = name
	meal.Calories = calories
	meal.Carbs = carbs
	meal.Protein = protein
	meal.Ingredients = ingredients
	meal.ImageUrl = src
	return meal, nil
}

func extractIngredientsFromModal(modal *rod.Element) (string, error) {
	ingredientsEl, err := modal.Element(".nutrition-text")
	if err != nil {
		return "", fmt.Errorf("can't parse ingredients %v", err)
	}
	ingredients, err := ingredientsEl.Text()
	if err != nil {
		return "", fmt.Errorf("can't parse ingredients %v", err)
	}
	return ingredients, nil
}

func extractImageFromModal(modal *rod.Element) (string, error) {
	imageEl, err := modal.Element(".modal-image")
	if err != nil {
		return "", fmt.Errorf("can't find image %v", err)
	}
	src, err := imageEl.Attribute("src")
	if err != nil {
		return "", fmt.Errorf("can't parse ingredients %v", err)
	}
	return "https://app.yumba.ca/" + *src, nil
}

func extractNameFromModal(modal *rod.Element) (string, error) {
	nameEl, err := modal.Element(".modal-name")
	if err != nil {
		return "", fmt.Errorf("can't parse name %v", err)
	}
	name, err := nameEl.Text()
	if err != nil {
		return "", fmt.Errorf("can't parse name %v", err)
	}
	return name, nil
}

func extractStatsFromModal(modal *rod.Element) (calories, carbs, protein int, error error) {
	modalStats, err := modal.Elements(".modals-stat")
	if err != nil {
		return 0, 0, 0, fmt.Errorf("can't parse stats %v", err)
	}
	for _, modalStat := range modalStats {
		pTag, err := modalStat.Element("p")
		if err != nil {
			return 0, 0, 0, fmt.Errorf("can't parse stats %v", err)
		}
		title, err := pTag.Text()
		if err != nil {
			return 0, 0, 0, fmt.Errorf("can't parse stats %v", err)
		}
		h4Tag, err := modalStat.Element("h4")
		if err != nil {
			return 0, 0, 0, fmt.Errorf("can't parse stats %v", err)
		}
		stat, err := h4Tag.Text()
		if err != nil {
			return 0, 0, 0, fmt.Errorf("can't parse stats %v", err)
		}
		statParsed, err := strconv.Atoi(stat)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("can't parse stats %v", err)
		}
		switch title {
		case "Calories":
			calories = statParsed
		case "Protein":
			protein = statParsed
		case "Carbs":
			carbs = statParsed
		default:
			panic(1)
		}
	}
	return calories, carbs, protein, nil
}

func addMealToNotion(meal *Meal) error {
	return nil
}
