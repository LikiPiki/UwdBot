package plugin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

const (
	weatherURL        = "http://api.openweathermap.org/data/2.5/group"
	weatherImageURL   = "http://openweathermap.org/img/wn/%s@2x.png"
	weatherUpdateTime = 20
)

var (
	// Map with cities, city ID from this json list - https://openweathermap.org/current
	weatherCities = map[string]City{
		"524894": City{
			"Москва", "Москве",
		},
		"536203": City{
			"Санкт-Петербург", "Санкт-Петербурге",
		},
		"1508291": City{
			"Челябинск", "Челябинске",
		},
		"706483": City{
			"Харьков", "Харькове",
		},
		"520555": City{
			"Нижний Новгород", "Нижнем Новгороде",
		},
		"1503940": City{
			"Кедровка", "В Кедровке",
		},
	}
)

// Filter - Filtering cities by city Name
func Filter(arr []string, cond func(string) bool) []string {
	result := []string{}
	for i := range arr {
		if cond(arr[i]) {
			result = append(result, arr[i])
		}
	}
	return result
}

type City struct {
	CityName   string
	InCityName string
}

type WeatherInlineReply struct {
	ID          string
	Content     string
	Description string
	ImageURL    string
}

// WeatherAPIResponse for weatherURL
type WeatherAPIResponse struct {
	Cnt  int `json:"cnt"`
	List []struct {
		Sys struct {
			Country  string `json:"country"`
			Timezone int    `json:"timezone"`
			Sunrise  int    `json:"sunrise"`
			Sunset   int    `json:"sunset"`
		} `json:"sys"`
		Weather []struct {
			ID          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Main struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			TempMin   float64 `json:"temp_min"`
			TempMax   float64 `json:"temp_max"`
			Pressure  int     `json:"pressure"`
			Humidity  int     `json:"humidity"`
		} `json:"main"`
		Visibility int `json:"visibility"`
		Wind       struct {
			Speed int `json:"speed"`
			Deg   int `json:"deg"`
		} `json:"wind"`
		Clouds struct {
			All int `json:"all"`
		} `json:"clouds"`
		Dt   int    `json:"dt"`
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"list"`
}

func (w *Weather) updateWeather(response *WeatherAPIResponse) error {
	req, err := http.NewRequest("GET", weatherURL, nil)

	if err != nil {
		return errors.Wrap(err, "cannot perform HTTP GET request")
	}

	weatherResp := WeatherAPIResponse{}

	q := req.URL.Query()
	q.Add("lang", "ru")
	q.Add("appid", w.weatherTOKEN)
	q.Add("units", "metric")

	citiesStr := ""
	citiesLen := len(weatherCities)
	i := 0
	for city := range weatherCities {
		citiesStr += city
		if citiesLen-1 != i {
			citiesStr += ","
		}
		i++
	}

	q.Add("id", citiesStr)
	req.URL.RawQuery = q.Encode()

	resp, err := http.Get(req.URL.String())
	if err != nil {
		return errors.Wrap(err, "cannot perform HTTP GET request")
	}
	defer resp.Body.Close()

	jsonCode, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "cannot read HTTP body")
	}

	err = json.Unmarshal(jsonCode, &weatherResp)
	if err != nil {
		return errors.Wrap(err, "cannot unmarshal json to weatherResp")
	}

	w.lastUpdateTime = time.Now()

	for i, cityWeather := range weatherResp.List {
		currentCity := weatherCities[strconv.Itoa(cityWeather.ID)]
		title := currentCity.CityName
		reply := WeatherInlineReply{
			ID: strconv.Itoa(i),
			Content: fmt.Sprintf(
				"В *%s* %.1f°, ощущается как %1.f°, %s. Ветер %d _м/с_, влажность %d%%",
				currentCity.InCityName,
				cityWeather.Main.Temp,
				cityWeather.Main.FeelsLike,
				cityWeather.Weather[0].Description,
				cityWeather.Wind.Speed,
				cityWeather.Main.Humidity,
			),
			Description: fmt.Sprintf(
				"В %s %.1f°, %s. Ветер %d м/с",
				currentCity.InCityName,
				cityWeather.Main.Temp,
				cityWeather.Weather[0].Description,
				cityWeather.Wind.Speed,
			),
			ImageURL: fmt.Sprintf(weatherImageURL, cityWeather.Weather[0].Icon),
		}
		w.replies[title] = reply
	}

	return nil
}
func (w *Weather) HandleWeatherInline(update *tgbotapi.Update) {
	if time.Since(w.lastUpdateTime).Minutes() >= weatherUpdateTime {
		weatherResp := WeatherAPIResponse{}
		err := w.updateWeather(&weatherResp)

		if err != nil {
			w.errors <- errors.Wrap(err, "cannot update weather")
		}
	}

	query := update.InlineQuery.Query

	items := make([]string, 0)
	for title := range w.replies {
		items = append(items, title)
	}

	filteredItems := Filter(items, func(item string) bool {
		return strings.Index(strings.ToLower(item), strings.ToLower(query)) >= 0
	})

	results := make(map[string]WeatherInlineReply)
	for _, item := range filteredItems {
		results[item] = w.replies[item]
	}

	var articles []interface{}
	i := 0
	for title, content := range results {
		msg := tgbotapi.NewInlineQueryResultArticleMarkdown(content.ID, title, content.Content)
		msg.Description = content.Description
		msg.ThumbURL = content.ImageURL
		msg.ThumbWidth = 32
		msg.ThumbHeight = 32

		articles = append(articles, msg)
		i++
	}

	inlineConfig := tgbotapi.InlineConfig{
		InlineQueryID: update.InlineQuery.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       articles,
	}

	if err := w.c.AnswerInlineQuery(&inlineConfig); err != nil {
		w.errors <- errors.Wrap(err, "cannot send inline reply")
	}
}
