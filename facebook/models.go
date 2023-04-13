package facebook

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

type EventCreationRequest struct {
	Data          []*Event `json:"data"`
	TestEventCode string   `json:"test_event_code,omitempty"`
}

type Event struct {
	EventName                    string               `json:"event_name"`
	EventTime                    int64                `json:"event_time"`
	UserData                     UserData             `json:"user_data,omitempty"`
	Contents                     []*ContentObject     `json:"contents,omitempty"`
	CustomData                   *CustomDataParameter `json:"custom_data,omitempty"`
	EventSourceUrl               string               `json:"event_source_url,omitempty"`
	OptOut                       bool                 `json:"opt_out,omitempty"`
	EventId                      string               `json:"event_id,omitempty"` // EventName + EventId must be unique
	ActionSource                 string               `json:"action_source" validate:"required"`
	DataProcessingOptions        *[]string            `json:"data_processing_options"`
	DataProcessingOptionsCountry *uint8               `json:"data_processing_options_country"`
	DataProcessingOptionsState   *uint8               `json:"data_processing_options_state"`
	AppData                      *AppData             `json:"app_data,omitempty"`
}

var validActionSources = map[string]struct{}{
	"email":            {},
	"website":          {},
	"app":              {},
	"phone_call":       {},
	"chat":             {},
	"physical_store":   {},
	"system_generated": {},
	"other":            {},
}

var eventConversion = map[string]string{
	"fb_mobile_initiated_checkout":    "InitiateCheckout",
	"fb_mobile_add_to_wishlist":       "AddToWishlist",
	"fb_mobile_complete_registration": "CompleteRegistration",
	"fb_mobile_purchase":              "Purchase",
	"fb_mobile_search":                "Search",
	"fb_mobile_content_view":          "ViewContent",
	"fb_mobile_add_to_cart":           "AddToCart",
	"fb_mobile_add_payment_info":      "AddPaymentInfo",
	"fb_mobile_spent_credits":         "SpentCredits",
}

func (e *Event) Validate() {
	if _, ok := validActionSources[e.ActionSource]; !ok {
		e.ActionSource = "other"
	}
	if en, ok := eventConversion[e.EventName]; ok {
		//Must be converted
		e.EventName = en
	}
	if e.EventName == "Purchase" {
		//Currency is required
		if e.CustomData == nil {
			e.CustomData = &CustomDataParameter{}
		}
		if e.CustomData.Currency == "" {
			e.CustomData.Currency = "EUR"
		}
		if e.CustomData.Value == 0 {
			//The event will not pass validation. We put a custom event name
			e.CustomData.Value = 42 //Fake value
		}
	}
}

type ExtInfoExpanded struct {
	Version                     string `json:"version"`
	PackageName                 string `json:"package_name"`
	ShortVersion                string `json:"short_version"`
	LongVersion                 string `json:"long_version"`
	OsVersion                   string `json:"os_version"`
	DeviceModel                 string `json:"device_model"`
	Locale                      string `json:"locale"`
	TimezoneAbbreviation        string `json:"timezone_abbreviation"` //es PDT
	Carrier                     string `json:"carrier"`
	ScreenWidth                 string `json:"screen_width"`
	ScreenHeight                string `json:"screen_height"`
	ScreenDensity               string `json:"screen_density"`
	CPUCores                    string `json:"cpu_cores"`
	ExternalStorageSize         string `json:"external_storage_size"`
	FreeSpaceOneExternalStorage string `json:"free_space_one_external_storage"`
	DeviceTimezone              string `json:"device_timezone"` //es USA/New York
}

func (e *Event) UnsetLimitedDataUse() {
	e.DataProcessingOptions = nil
	e.DataProcessingOptionsCountry = nil
	e.DataProcessingOptionsState = nil
}

func (e *Event) SetLimitedDataUse(how bool) {
	dpo := make([]string, 0)
	if !how {
		e.DataProcessingOptions = &dpo
		e.DataProcessingOptionsCountry = nil
		e.DataProcessingOptionsState = nil
		return
	}
	dpo = append(dpo, "LDU")
	e.DataProcessingOptions = &dpo
	allCountries := uint8(0)
	e.DataProcessingOptionsCountry = &allCountries
	e.DataProcessingOptionsState = &allCountries
}

type AppData struct {
	AdvertiserTrackingEnabled  int8     `json:"advertiser_tracking_enabled"`
	ApplicationTrackingEnabled int8     `json:"application_tracking_enabled"`
	Extinfo                    []string `json:"extinfo"`
	InstallReferrer            string   `json:"install_referrer,omitempty"`
	InstallerPackage           string   `json:"installer_package,omitempty"`
	UrlSchemes                 []string `json:"url_schemes,omitempty"`
	WindowsAttributionId       string   `json:"windows_attribution_id,omitempty"`
}

func (e *ExtInfoExpanded) SetAndroid() {
	e.Version = "a2"
}

func (e *ExtInfoExpanded) SetIOS() {
	e.Version = "i2"
}

func (d *AppData) FillExtInfo(in ExtInfoExpanded) {
	d.Extinfo = []string{
		in.Version,
		in.PackageName,
		in.ShortVersion,
		in.LongVersion,
		in.OsVersion,
		in.DeviceModel,
		in.Locale,
		in.TimezoneAbbreviation,
		in.Carrier,
		in.ScreenWidth,
		in.ScreenHeight,
		in.ScreenDensity,
		in.CPUCores,
		in.ExternalStorageSize,
		in.FreeSpaceOneExternalStorage,
		in.DeviceTimezone,
	}
}

type UserData struct {
	Emails          []string `json:"em,omitempty"`
	Phones          []string `json:"ph,omitempty"`
	FirstName       []string `json:"fn,omitempty"`
	LastName        []string `json:"ln,omitempty"`
	BirthDate       string   `json:"db,omitempty"`
	Genre           string   `json:"ge,omitempty"`
	City            []string `json:"city,omitempty"`
	ClientIpAddress string   `json:"client_ip_address,omitempty"`
	ClientUserAgent string   `json:"client_user_agent,omitempty"`
	Fbc             string   `json:"fbc,omitempty"`
	Fbp             string   `json:"fbp,omitempty"`
	ExternalId      string   `json:"external_id,omitempty"`
	SubscriptionId  string   `json:"subscription_id,omitempty"`
	LoginId         int64    `json:"facebook_login_id,omitempty"`
	LeadId          int64    `json:"lead_id,omitempty"`
}

func (u *UserData) AddCity(name string) {
	//Hash and append
	u.City = append(u.City, Hash(name))
}

func (u *UserData) AddFirstName(name string) {
	//Hash and append
	u.FirstName = append(u.FirstName, Hash(name))
}

func (u *UserData) AddLastName(name string) {
	//Hash and append
	u.LastName = append(u.LastName, Hash(name))
}

func (u *UserData) AddEmail(email string) {
	//Hash and append
	u.Emails = append(u.Emails, Hash(email))
}

func (u *UserData) AddPhone(phone string) {
	//Hash and append
	u.Phones = append(u.Phones, Hash(phone))
}

func Hash(s string) string {
	s = strings.ToLower(s)
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

type CustomDataParameter struct {
	ContentCategory  string           `json:"content_category,omitempty"`
	ContentIds       []string         `json:"content_ids,omitempty"`
	ContentName      string           `json:"content_name,omitempty"`
	ContentType      string           `json:"content_type,omitempty"`
	Contents         []*ContentObject `json:"contents,omitempty"`
	Currency         string           `json:"currency,omitempty"`
	DeliveryCategory string           `json:"delivery_category,omitempty"`
	NumItems         string           `json:"num_items,omitempty"` //Use *only* with InitiateCheckout events
	OrderId          string           `json:"order_id,omitempty"`
	PredictedLtv     float64          `json:"predicted_ltv,omitempty"`
	SearchString     string           `json:"search_string,omitempty"`
	Status           string           `json:"status,omitempty"`
	Value            float64          `json:"value,omitempty"`
}

type ContentObject struct {
	Id               string  `json:"id,omitempty"`
	Quantity         int     `json:"quantity,omitempty"`
	ItemPrice        float64 `json:"item_price,omitempty"`
	DeliveryCategory string  `json:"delivery_category,omitempty"`
}
