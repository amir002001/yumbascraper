package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
		_, err := addMealToNotion(&meals[i])
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
			return 0, 0, 0, fmt.Errorf("unexpected title")
		}
	}
	return calories, carbs, protein, nil
}

func addMealToNotion(meal *Meal) (bool, error) {
	createPayload, err := prepareCreatePayload(meal)
	if err != nil {
		return false, err
	}
	queryPayload := prepareQueryPaylaod(meal)
	secret := os.Getenv("NOTION_SECRET")
	if secret == "" {
		return false, fmt.Errorf("notion secret not set")
	}
	exists, err := checkMealExistsOnNotion(secret, &queryPayload)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}
	err = createMealOnNotion(secret, &createPayload)
	if err != nil {
		return false, err
	}
	return true, nil
}

func checkMealExistsOnNotion(secret string, payload *QueryPayload) (bool, error) {
	url := "https://api.notion.com/v1/databases/75768c18989642edabf16508ee4233fa/query"
	method := "POST"

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}
	reader := bytes.NewReader(jsonPayload)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Notion-Version", "2022-02-22")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secret))

	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	fmt.Println(body)
	return false, nil
}

func createMealOnNotion(secret string, payload *CreatePayload) error {
	url := "https://api.notion.com/v1/pages/"
	method := "POST"
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(jsonPayload)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Notion-Version", "2022-02-22")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secret))

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	bodyStr := string(body)
	if err != nil {
		return err
	}
	if res.StatusCode < 400 {
		fmt.Println(bodyStr)
		return nil
	}
	return fmt.Errorf(bodyStr)
}

func prepareQueryPaylaod(meal *Meal) QueryPayload {
	payload := QueryPayload{
		Filter: Filter{And: []And{{Property: "Name", RichText: QueryRichText{Equals: meal.Name}}}},
	}
	return payload
}

func prepareCreatePayload(meal *Meal) (CreatePayload, error) {
	databaseId := os.Getenv("DATABASE_ID")
	if databaseId == "" {
		return CreatePayload{}, fmt.Errorf("database id not set")
	}
	parent := Parent{DatabaseID: databaseId}
	external := External{URL: meal.ImageUrl}
	cover := Cover{Type: "external", External: external}
	properties := Properties{
		Name:     Name{Title: []Title{{Text: Text{Content: meal.Name}}}},
		Carbs:    Carbs{Type: "number", Number: meal.Carbs},
		Calories: Calories{Type: "number", Number: meal.Calories},
		Protein:  Protein{Type: "number", Number: meal.Protein},
		Ingredients: Ingredients{
			Type:     "rich_text",
			RichText: []RichText{{Text: Text{Content: meal.Ingredients}}},
		},
	}
	payload := CreatePayload{Parent: parent, Cover: cover, Properties: properties}
	return payload, nil
}

type CreatePayload struct {
	Parent     Parent     `json:"parent"`
	Cover      Cover      `json:"cover"`
	Properties Properties `json:"properties"`
}
type Parent struct {
	DatabaseID string `json:"database_id"`
}
type External struct {
	URL string `json:"url"`
}
type Cover struct {
	Type     string   `json:"type"`
	External External `json:"external"`
}
type Text struct {
	Content string `json:"content"`
}
type Title struct {
	Text Text `json:"text"`
}
type Name struct {
	Title []Title `json:"title"`
}
type Calories struct {
	Type   string `json:"type"`
	Number int    `json:"number"`
}
type Protein struct {
	Type   string `json:"type"`
	Number int    `json:"number"`
}
type Carbs struct {
	Type   string `json:"type"`
	Number int    `json:"number"`
}
type RichText struct {
	Text Text `json:"text"`
}
type Ingredients struct {
	Type     string     `json:"type"`
	RichText []RichText `json:"rich_text"`
}
type Properties struct {
	Name        Name        `json:"Name"`
	Calories    Calories    `json:"Calories"`
	Protein     Protein     `json:"Protein"`
	Carbs       Carbs       `json:"Carbs"`
	Ingredients Ingredients `json:"Ingredients"`
}

type QueryPayload struct {
	Filter Filter `json:"filter"`
}
type QueryRichText struct {
	Equals string `json:"equals"`
}
type And struct {
	Property string        `json:"property"`
	RichText QueryRichText `json:"rich_text"`
}
type Filter struct {
	And []And `json:"and"`
}
