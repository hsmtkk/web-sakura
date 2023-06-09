package function

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"golang.org/x/net/publicsuffix"
)

const WEB_SAKURA = "parents.cloud-sakura.net"
const LOGIN_PATH = "/pages/accounts/login.php"
const REGIST_PATH = "/pages/contact-book/regist-api.php"

func init() {
	functions.HTTP("EntryPoint", EntryPoint)
}

func EntryPoint(w http.ResponseWriter, r *http.Request) {
	log.Print("begin")
	if err := entryPointInner(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	log.Print("end")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func entryPointInner(ctx context.Context) error {
	account := neededEnvVar("ACCOUNT")
	password := neededEnvVar("PASSWORD")
	childID := neededEnvVar("CHILD_ID")
	projectID := neededEnvVar("PROJECT_ID")
	saveDataCollection := neededEnvVar("SAVE_DATA_COLLECTION")
	saveDataDocument := neededEnvVar("SAVE_DATA_DOCUMENT")

	hdl, err := newHandler()
	if err != nil {
		return err
	}

	if err := hdl.login(account, password); err != nil {
		return err
	}
	if err := hdl.regist(ctx, projectID, saveDataCollection, saveDataDocument, childID); err != nil {
		return err
	}
	return nil
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
	postURL := fmt.Sprintf("https://%s%s", WEB_SAKURA, LOGIN_PATH)
	form := url.Values{"account": []string{account}, "password": []string{password}}
	if err := h.httpPost(postURL, form); err != nil {
		return err
	}
	return nil
}

func (h *handler) regist(ctx context.Context, projectID, saveDataCollection, saveDataDocument, childID string) error {
	postURL := fmt.Sprintf("https://%s%s", WEB_SAKURA, REGIST_PATH)
	registData, err := newRegistData(ctx, projectID, saveDataCollection, saveDataDocument)
	if err != nil {
		return err
	}
	saveDataBytes, err := json.Marshal(registData)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON; %w", err)
	}
	saveData := string(saveDataBytes)
	log.Print(saveData)
	form := url.Values{"save_data": []string{saveData}, "child_id": []string{childID}}
	if err := h.httpPost(postURL, form); err != nil {
		return err
	}
	return nil
}

func (h *handler) httpPost(postURL string, form url.Values) error {
	resp, err := h.client.PostForm(postURL, form)
	if err != nil {
		return fmt.Errorf("failed to send HTTP POST; %w", err)
	}
	defer resp.Body.Close()
	respDump, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return fmt.Errorf("failed to dump HTTP response; %w", err)
	}
	log.Print(string(respDump))
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("got HTTP error code; %d; %s", resp.StatusCode, resp.Status)
	}
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

func newRegistData(ctx context.Context, projectID, saveDataCollection, saveDataDocument string) (registData, error) {
	var regData registData
	clt, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return regData, fmt.Errorf("failed to make Firestore client; %w", err)
	}
	defer clt.Close()
	snap, err := clt.Collection(saveDataCollection).Doc(saveDataDocument).Get(ctx)
	if err != nil {
		return regData, fmt.Errorf("failed to get Firestore document; %w", err)
	}
	if err := snap.DataTo(&regData); err != nil {
		return regData, fmt.Errorf("failed to convert data; %w", err)
	}

	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return regData, fmt.Errorf("failed to set timezone; %w", err)
	}
	time.Local = location
	now := time.Now()
	regData.DateMonth = now.Format("01")
	regData.DateDay = now.Format("02")

	thermos := strings.Split(regData.Thermometry, ".")
	regData.ThermometryPre = thermos[0]
	regData.ThermometryAfter = thermos[1]

	return regData, nil

	/*
		return registData{
			DateMonth:                month,
			DateDay:                  day,
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
		}, nil
	*/
}
