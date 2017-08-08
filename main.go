package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

type history struct {
	Count   int    `json:"count"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
	History []trip `json:"history"`
}

type trip struct {
	Distance    float64   `json:"distance"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	RequestTime time.Time `json:"request_time"`
	Status      string    `json:"status"`
}

type stats struct {
	Count        int           `json:"count"`
	TotalWaiting time.Duration `json:"totalWaiting"`
	TotalOnTrip  time.Duration `json:"totalOnTrip"`
}

func main() {
	e := echo.New()
	e.GET("/api/oauth/callback", func(c echo.Context) error {
		resp, err := http.PostForm("https://login.uber.com/oauth/v2/token", url.Values{
			"client_secret": {"GET FROM UBER.COM"},
			"client_id":     {"GET FROM UBER.COM"},
			"redirect_uri":  {"http://localhost:3000/api/oauth/callback"},
			"grant_type":    {"authorization_code"},
			"code":          {c.QueryParam("code")},
		})
		if err != nil {
			fmt.Println(fmt.Errorf("Error posting request: %+v", err))
			return err
		}

		decoder := json.NewDecoder(resp.Body)
		parsed := map[string]interface{}{}
		err = decoder.Decode(&parsed)
		if err != nil {
			fmt.Println(fmt.Errorf("Error decoding response: %+v", err))
			return err
		}

		accessToken := parsed["access_token"].(string)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"accessToken": accessToken,
		})
		tokenString, err := token.SignedString([]byte("secret"))
		if err != nil {
			fmt.Println(fmt.Errorf("Error creating token string: %+v", err))
			return err
		}
		c.SetCookie(&http.Cookie{
			Name:  "jwt",
			Value: tokenString,
			Path:  "/",
		})
		return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3000")
	})

	e.GET("/api/auth", func(c echo.Context) error {
		jwtCookie, err := c.Cookie("jwt")
		if err != nil {
			fmt.Println(fmt.Errorf("Error getting jwt error: %+v", err))
			return err
		}

		tokenString := jwtCookie.Value
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				fmt.Println(fmt.Errorf("Unexpected signing method"))
				return nil, nil
			}

			return []byte("secret"), nil
		})

		if _, ok := token.Claims.(jwt.MapClaims); !(ok && token.Valid) {
			fmt.Println(fmt.Errorf("Error getting claims: %+v", err))
			return err
		}
		return c.NoContent(http.StatusOK)
	})

	e.GET("/api/stats", func(c echo.Context) error {
		jwtCookie, err := c.Cookie("jwt")
		if err != nil {
			fmt.Println(fmt.Errorf("Error getting jwt error: %+v", err))
			return err
		}

		tokenString := jwtCookie.Value
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				fmt.Println(fmt.Errorf("Unexpected signing method"))
				return nil, nil
			}

			return []byte("secret"), nil
		})

		accessToken := ""
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			accessToken = claims["accessToken"].(string)
		} else {
			fmt.Println(fmt.Errorf("Error getting claims: %+v", err))
			return err
		}

		s := stats{}
		offset := 0
		limit := 50
		count := 0
		first := true
		for first || offset < count {
			first = false
			req, err := http.NewRequest("GET", "https://api.uber.com/v1.2/history?limit="+strconv.Itoa(limit)+"&offset="+strconv.Itoa(offset), nil)
			if err != nil {
				fmt.Println(fmt.Errorf("Error creating uber api request: %+v", err))
				return err
			}
			req.Header.Add("Authorization", "Bearer "+accessToken)
			req.Header.Add("Accept-Language", "en_US")
			req.Header.Add("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println(fmt.Errorf("Error on request to uber api: %+v", err))
				return err
			}

			defer resp.Body.Close()
			parsed := history{}
			err = json.NewDecoder(resp.Body).Decode(&parsed)
			if err != nil {
				fmt.Println(fmt.Errorf("Error decoding response: %+v", err))
				return err
			}

			for _, t := range parsed.History {
				s.TotalOnTrip += (t.EndTime.Sub(t.StartTime))
				s.TotalWaiting += (t.StartTime.Sub(t.RequestTime))
			}

			offset += limit
			count = parsed.Count
			s.Count = count
		}

		return c.JSON(http.StatusOK, s)
	})
	e.Logger.Fatal(e.Start(":8080"))
}

func (t *trip) UnmarshalJSON(data []byte) error {
	type Alias trip
	aux := &struct {
		StartTime   int64 `json:"start_time"`
		EndTime     int64 `json:"end_time"`
		RequestTime int64 `json:"request_time"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	t.StartTime = time.Unix(aux.StartTime, 0)
	t.EndTime = time.Unix(aux.EndTime, 0)
	t.RequestTime = time.Unix(aux.RequestTime, 0)
	return nil
}
