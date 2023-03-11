package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"

	"golang.org/x/net/publicsuffix"
)

const WEB_SAKURA = "parents.cloud-sakura.net"
const LOGIN_PATH = "/pages/accounts/login.php"
const CALENDAR_PATH = "/pages/calendar/index.php"
const REGIST_PATH = "/pages/contact-book/regist-api.php"

func main() {
	account := neededEnvVar("ACCOUNT")
	password := neededEnvVar("PASSWORD")
	childID := neededEnvVar("CHILD_ID")
	hdl, err := newHandler()
	if err != nil {
		log.Fatal(err)
	}
	if err := hdl.login(account, password); err != nil {
		log.Fatal(err)
	}
	/*
		if err := hdl.getCalendar(); err != nil {
			log.Fatal(err)
		}
	*/
	if err := hdl.regist(childID); err != nil {
		log.Fatal(err)
	}
}

func neededEnvVar(name string) string {
	val := os.Getenv(name)
	if val == "" {
		log.Fatalf("%s env var must be defined", name)
	}
	return val
}

type handler struct {
	client *http.Client
}

func newHandler() (*handler, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("failed to make cookie jar; %w", err)
	}
	client := &http.Client{
		Jar: jar,
	}
	return &handler{client}, nil
}

func (h *handler) login(account, password string) error {
	loginURL := fmt.Sprintf("https://%s%s", WEB_SAKURA, LOGIN_PATH)
	values := url.Values{"account": []string{account}, "password": []string{password}}
	resp, err := h.client.PostForm(loginURL, values)
	if err != nil {
		return fmt.Errorf("failed to send HTTP POST; %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("got HTTP error code; %d; %s", resp.StatusCode, resp.Status)
	}
	return nil
}

func (h *handler) getCalendar() error {
	calendarURL := fmt.Sprintf("https://%s%s", WEB_SAKURA, CALENDAR_PATH)
	resp, err := h.client.Get(calendarURL)
	if err != nil {
		return fmt.Errorf("failed to send HTTP GET; %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("got HTTP error code; %d; %s", resp.StatusCode, resp.Status)
	}
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return fmt.Errorf("failed to dump HTTP response; %w", err)
	}
	log.Print(string(respDump))
	return nil
}

type registData struct {
	DateMonth                string `json:"date_month"`
	DateDay                  string `json:"date_day"`
	MoodyLastnight           string `json:"moody_lastnight"`
	DefecationLastnightCount string `json:"defecation_lastnight_count"`
	DefecationLastnight      string `json:"defecation_lastnight"`
	DinnerHour               string `json:"dinner_hour"`
	DinnerMinus              string `json:"dinner_minus"`
	DinnerComment            string `json:"dinner_comment"`
	Bathing                  string `json:"bathing"`
	SleepHour                string `json:"sleep_hour"`
	SleepMinus               string `json:"sleep_minus"`
	WakeupHour               string `json:"wakeup_hour"`
	WakeupMinus              string `json:"wakeup_minus"`
	MoodyMorning             string `json:"moody_morning"`
	DefecationMorningCount   string `json:"defecation_morning_count"`
	DefecationMorning        string `json:"defecation_morning"`
	BreakfastHour            string `json:"breakfirst_hour"`
	BreakfastMinus           string `json:"breakfirst_minus"`
	BreakfastComment         string `json:"breakfirst_comment"`
	ThermometryHour          string `json:"thermometry_hour"`
	ThermometryMinus         string `json:"thermometry_minus"`
	ThermometryPre           string `json:"thermometry_pre"`
	ThermometryAfter         string `json:"thermometry_after"`
	Swimming                 string `json:"swimming"`
	Appearance               string `json:"appearance"`
	Message                  string `json:"message"`
	RelationshipID           string `json:"relationship_id"`
	PickupHour               string `json:"pickup_hour"`
	PickupMinus              string `json:"pickup_minus"`
	SaveState                string `json:"save_state"`
	Thermometry              string `json:"thermometry"`
}

func newRegistData() registData {
	return registData{
		DateMonth:                "03",
		DateDay:                  "10",
		MoodyLastnight:           "1",
		DefecationLastnightCount: "0",
		DefecationLastnight:      "0",
		DinnerHour:               "19",
		DinnerMinus:              "00",
		DinnerComment:            "食べました。",
		Bathing:                  "2",
		SleepHour:                "20",
		SleepMinus:               "30",
		WakeupHour:               "06",
		WakeupMinus:              "30",
		MoodyMorning:             "1",
		DefecationMorningCount:   "0",
		DefecationMorning:        "0",
		BreakfastHour:            "07",
		BreakfastMinus:           "00",
		BreakfastComment:         "食べました。",
		ThermometryHour:          "07",
		ThermometryMinus:         "30",
		ThermometryPre:           "36",
		ThermometryAfter:         "7",
		Swimming:                 "1", // 1なし 2あり
		Appearance:               "いつも通りです。",
		Message:                  "",
		RelationshipID:           "父",
		PickupHour:               "17",
		PickupMinus:              "50",
		SaveState:                "2",
		Thermometry:              "36.7",
	}
}

func (h *handler) regist(childID string) error {
	registURL := fmt.Sprintf("https://%s%s", WEB_SAKURA, REGIST_PATH)
	saveData, err := json.Marshal(newRegistData())
	if err != nil {
		return fmt.Errorf("failed to marshal JSON; %w", err)
	}
	log.Print(string(saveData))
	values := url.Values{"save_data": []string{string(saveData)}, "child_id": []string{childID}}
	resp, err := h.client.PostForm(registURL, values)
	if err != nil {
		return fmt.Errorf("failed to send HTTP POST; %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("got HTTP error code; %d; %s", resp.StatusCode, resp.Status)
	}
	return nil
}
