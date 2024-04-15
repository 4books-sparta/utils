package intercom

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/intercom/intercom-go.v2"

	"github.com/4books-sparta/utils"
)

const (
	perPage                       = 50
	BooksStartedKey               = "books_started"
	BooksCompletedKey             = "books_completed"
	ErrorUserNotFound             = "intercom-user-not-found"
	ErrorIntercomEventUnsupported = "ErrorIntercomEventUnsupported"
)

// map to translate inner fields to intercom fields
var customFieldsMap = map[string]string{
	"user_language": "language",
}

// map of not allowed fields to be transmitted to intercom
var blacklistFieldsMap = map[string]interface{}{}

type Client struct {
	ic      *intercom.Client
	Verbose bool
}

func New(id, key string) *Client {
	return &Client{
		ic: intercom.NewClient(key, ""),
	}
}

func isNotFound(err error) bool {
	herr, ok := err.(intercom.IntercomError)
	return ok && herr.GetCode() == "not_found"
}

func preFill(u *User, remote *intercom.User) {
	//TODO understand if it's better to always update
	if len(remote.UserID) <= 0 {
		remote.UserID = u.Id
	}

	if u.Email != "" {
		remote.Email = u.Email
	}
	if u.FullName != "" {
		remote.Name = u.FullName
	}
	if u.CreatedAt != nil && !u.CreatedAt.IsZero() {
		remote.SignedUpAt = u.CreatedAt.Unix()
	}
}

func (c *Client) TraceHTTP(val bool) {
	c.ic.Option(intercom.TraceHTTP(val))
}

func (c *Client) Log(msg string) {
	if c.Verbose {
		fmt.Println(time.Now().UnixMilli(), " => ", msg)
	}
}

func (c *Client) Dump(title string, msg interface{}) {
	if c.Verbose {
		utils.PrintVarDump(title, msg)
	}
}

func (c *Client) GetUserByEmail(e string) (*intercom.User, error) {
	c.Log("GetUserByEmail: " + e)
	existing, err := c.ic.Users.FindByEmail(e)
	if err != nil {
		c.Log("GetUserByEmail error: " + err.Error())
		return nil, err
	}
	c.Log("GetUserByEmail found:" + existing.UserID)
	return &existing, nil
}

func (c *Client) ListAdmins() (*intercom.AdminList, error) {
	existing, err := c.ic.Admins.List()
	if err != nil {
		return nil, err
	}

	return &existing, nil
}

func (c *Client) FindAdminByEmail(email string) (*intercom.Admin, error) {
	admins, err := c.ListAdmins()
	if err != nil {
		return nil, err
	}
	if admins == nil {
		return nil, nil
	}
	for _, a := range admins.Admins {
		if a.Email == email {
			return &a, nil
		}
	}

	return nil, nil
}

func (c *Client) CreateNewMessage(emailFrom, emailTo, subject, body string) (*intercom.MessageRequest, *intercom.User, error) {
	//fmt.Println("Email from: <",emailFrom,"> to: <",emailTo,">")
	icUser, err := c.GetUserByEmail(emailTo)
	if err != nil {
		return nil, nil, err
	}

	admin, err := c.FindAdminByEmail(emailFrom)
	if err != nil {
		return nil, nil, err
	}
	if admin == nil {
		return nil, nil, errors.New("admin-not-found")
	}

	message := intercom.NewEmailMessage(intercom.NO_TEMPLATE, admin, icUser, subject, body)
	return &message, icUser, nil
}

func (c *Client) UpdateReferralInfo(email string, uid uint32, refId uint32) error {
	custom := map[string]interface{}{
		"referred_by": refId,
	}
	err := c.save(&User{
		Id:    strconv.Itoa(int(uid)),
		Email: email,
	}, custom)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateReferrerFriendsCount(email string, uid uint32) error {
	return c.UpdateCustomIncrementalUserValue(email, uid, "friends_signed_up")
}

func (c *Client) UpdateCustomIncrementalUserValue(email string, uid uint32, key string) error {
	user, err := c.matchUser(&User{
		Id:    strconv.Itoa(int(uid)),
		Email: email,
	})
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New(ErrorUserNotFound)
	}

	var count int
	old, ok := user.CustomAttributes[key].(float64)
	if !ok {
		count = 0
	} else {
		count = int(old)
	}

	count++

	return c.save(&User{
		Id:    strconv.Itoa(int(uid)),
		Email: email,
	}, map[string]interface{}{
		key: count,
	})
}

func (c *Client) UpdateReferrerSubscribedFriendsCount(email string, uid uint32) error {
	return c.UpdateCustomIncrementalUserValue(email, uid, "friends_subscribed")
}

func (c *Client) SendMessage(msg *intercom.MessageRequest) (intercom.MessageResponse, error) {
	return c.ic.Messages.Save(msg)
}

// Tries to match an internal user to an intercom's one
// given the convoluted logic used by intercom to do so
func (c *Client) matchUser(u *User) (*intercom.User, error) {
	var existing intercom.User
	var err error

	if u.Email != "" {
		c.Log("matchUser by email" + u.Email)
		existing, err = c.ic.Users.FindByEmail(u.Email)
		if err == nil && existing.Email != "" {
			// no error retrieving it so we have a user
			c.Dump("Matched by email: ", existing)
			return &existing, nil
		} else {
			c.Dump("Error is", err)
			c.Dump("Existing", existing)
		}
	}

	if existing.Email == "" || isNotFound(err) {
		// the user can not be found via email so try with our id
		c.Log("matchUser by id " + u.Id)
		existing, err = c.ic.Users.FindByUserID(u.Id)
		if isNotFound(err) {
			// the user was not found even with an id  so we can
			// safely say that our user does not exist on intercom
			return nil, nil
		} else if err != nil {
			// another error occurred while querying using the id
			// so return it
			return nil, err
		}

		// we have found a user using its id
		c.Dump("Matched: ", existing)
		if u.Id == existing.UserID {
			return &existing, nil
		}
		c.Log("mismatch-userId-found-but-different: " + u.Id + "::" + existing.ID)
	}

	// there was an unhandled error querying for the user
	// using its email
	return nil, err
}

func (c *Client) save(u *User, custom map[string]interface{}) error {
	c.Dump("saving from...", u)
	c.Dump("custom...", custom)
	user, err := c.matchUser(u)
	if err != nil {
		return err
	}

	if user == nil {
		c.Log("intercom-user-not-matched")
		//The user cant be created/updated if it's not verified
		if !(u.Verified || u.UserVerified) {
			c.Log("cant-create-new-user-not-verified: " + u.Email)
			return nil
		}

		// no user matching the given email
		user = &intercom.User{
			CustomAttributes: make(map[string]interface{}),
		}
	}

	if user.CustomAttributes == nil {
		user.CustomAttributes = make(map[string]interface{})
	}

	//Another check
	oldV := false
	oldEV := false

	if v, ok := user.CustomAttributes["user_verified"]; ok {
		if vb, ok := v.(bool); ok {
			oldV = vb
		}
		if vb, ok := v.(string); ok {
			oldV = vb == "true"
		}
	}
	if v, ok := user.CustomAttributes["email_verified"]; ok {
		if vb, ok := v.(bool); ok {
			oldEV = vb
		}
		if vb, ok := v.(string); ok {
			oldEV = vb == "true"
		}
	}

	if !(oldV || oldEV || u.Verified || u.UserVerified) {
		c.Log("cant-save-new-user-not-verified: " + u.Email)
		return nil
	}
	// ensure the base fields
	preFill(u, user)

	if u.FullName == "deleted" {
		deleted := true
		user.UnsubscribedFromEmails = &deleted
	}

	for k, v := range custom {
		user.CustomAttributes[k] = v
	}

	//Override attributes come with the request
	if u.CustomFields != nil && len(u.CustomFields) > 0 {
		for k, v := range u.CustomFields {
			user.CustomAttributes[k] = v
		}
	}

	if len(user.UserID) <= 0 {
		errStr := fmt.Sprintf("Cannot save with no ID: %s", u.Email)
		return errors.New(errStr)
	}

	c.Dump("Save after fill", user)

	res, err := c.ic.Users.Save(user)
	if err != nil {
		fmt.Println("Error saving to intercom: ", err)
	}

	u.RawUser = &res
	c.Dump("Saved", res)

	return err
}

func (c *Client) SaveUser(u *User) error {
	if u == nil {
		return nil
	}
	custom := make(map[string]interface{})
	if u.Subscription != nil && u.Subscription.Provider != "" {
		custom["subscription_status"] = u.Subscription.Status
		custom["provider_name"] = u.Subscription.Provider
		custom["plan_name"] = u.Subscription.Plan

		if len(u.Subscription.Products) > 0 {
			custom["subscription_products"] = u.Subscription.Products
		} else {
			custom["subscription_products"] = "none"
		}

		if len(u.Subscription.Company) > 0 {
			custom["company"] = u.Subscription.Company
		}

		if u.Subscription.Expiry.IsZero() {
			custom["subscription_expiry_at"] = ""
		} else {
			custom["subscription_expiry_at"] = u.Subscription.Expiry.Unix()
		}

		if u.Subscription.LastDisabledAt.IsZero() {
			custom["cf_renewal_disabled_at"] = ""
		} else {
			custom["cf_renewal_disabled_at"] = u.Subscription.LastDisabledAt.Unix()
		}

		if u.Subscription.LastEnabledAt.IsZero() {
			custom["cf_renewal_reenabled_at"] = ""
		} else {
			custom["cf_renewal_reenabled_at"] = u.Subscription.LastEnabledAt.Unix()
		}

		if u.Subscription.CreatedAt.IsZero() {
			custom["subscription_start_at"] = ""
		} else {
			custom["subscription_start_at"] = u.Subscription.CreatedAt.Unix()
		}

		if u.Subscription.CancelledAt.IsZero() {
			custom["subscription_cancelled_at"] = ""
		} else {
			custom["subscription_cancelled_at"] = u.Subscription.CancelledAt.Unix()
		}

		if u.Subscription.TrialStart.IsZero() {
			custom["trial_start_at"] = ""
		} else {
			custom["trial_start_at"] = u.Subscription.TrialStart.Unix()
		}

		if u.Subscription.TrialEnd.IsZero() {
			custom["trial_end_at"] = ""
		} else {
			custom["trial_end_at"] = u.Subscription.TrialEnd.Unix()
		}

	}
	if u.Verified {
		custom["email_verified"] = u.Verified
	}
	if u.UserVerified {
		custom["user_verified"] = u.UserVerified
	}

	if u.ABTestVariant != nil {
		custom["last_ab"] = u.ABTestVariant.Text
		custom["last_ab_date"] = u.ABTestVariant.CreatedAt.Unix()
	}

	if u.CustomFields != nil && len(u.CustomFields) > 0 {
		for k, v := range u.CustomFields {
			//Check blacklisted
			if _, ok := blacklistFieldsMap[k]; ok {
				continue
			}

			//Translate if mapped
			if _, ok := customFieldsMap[k]; ok {
				k = customFieldsMap[k]
			}
			//Add to custom fields
			custom[k] = v
		}
	}

	return c.save(u, custom)
}

type ProgressCount struct {
	Started   uint16
	Completed uint16
}

func (c *Client) GetProgressCount(u *User) (*ProgressCount, error) {
	iu, err := c.matchUser(&User{
		Id:    u.Id,
		Email: u.Email,
	})
	if err != nil {
		return nil, err
	}
	if iu == nil {
		return nil, errors.New(ErrorUserNotFound)
	}

	start, ok := iu.CustomAttributes[string(BooksStartedKey)].(float64)
	if !ok {
		start = 0
	}

	complete, ok := iu.CustomAttributes[string(BooksCompletedKey)].(float64)
	if !ok {
		complete = 0
	}

	pc := ProgressCount{
		Started:   uint16(start),
		Completed: uint16(complete),
	}
	return &pc, nil
}

var FunnelSupported = map[string]struct{}{}

func (c *Client) SaveUserFunnel(u *User, key string, val uint32) error {
	custom := make(map[string]interface{})
	custom[key] = val
	if _, ok := FunnelSupported[key]; ok {
		return c.save(u, custom)
	}
	return errors.New(ErrorIntercomEventUnsupported)
}

func (c *Client) UpdateCustomField(u *User, key string, val uint32) error {
	custom := make(map[string]interface{})
	custom[key] = val
	return c.save(u, custom)
}

func (c *Client) SaveCompany(co intercom.Company) (intercom.Company, error) {
	return c.ic.Companies.Save(&co)
}

func (c *Client) GetCompanyByExternalId(id string) (intercom.Company, error) {
	return c.ic.Companies.FindByCompanyID(id)
}

func (c *Client) SaveProgress(u *User) error {
	custom := make(map[string]interface{})

	// update started and completed data
	custom[BooksCompletedKey] = u.BooksCompleted
	custom[BooksStartedKey] = u.BooksStarted
	if u.LastStarted != nil {
		custom["last_book_started_slug"] = u.LastStarted.Slug
		custom["last_book_started_title"] = u.LastStarted.Title
		custom["last_book_started_at"] = u.LastStarted.At.Unix()
		// reset data
		if u.LastStarted.At.IsZero() {
			custom["last_book_started_at"] = ""
		}
	}
	if u.LastCompleted != nil {
		custom["last_book_completed_slug"] = u.LastCompleted.Slug
		custom["last_book_completed_title"] = u.LastCompleted.Title
		custom["last_book_completed_at"] = u.LastCompleted.At.Unix()
		if u.LastCompleted.At.IsZero() {
			// reset data
			custom["last_book_completed_at"] = ""
		}
	}

	return c.save(u, custom)
}

func (c *Client) SaveInterest(u *User, i string) error {
	custom := make(map[string]interface{})
	custom["interest"] = i
	return c.save(u, custom)
}

func (c *Client) SaveScore(u *User) error {
	custom := make(map[string]interface{})
	if u.Score != nil {
		custom["nps"] = *u.Score
	}

	return c.save(u, custom)
}

func (c *Client) ListUsers(page int) ([]intercom.User, error) {
	res, err := c.ic.Users.List(intercom.PageParams{
		Page:    int64(page),
		PerPage: perPage,
	})
	if err != nil {
		return nil, err
	}

	return res.Users, nil
}

func (c *Client) DeleteUser(id string) error {
	_, err := c.ic.Users.Delete(id)
	return err
}

type UserCompany struct {
	UserId    string
	CompanyId string
}

func (c *Client) AttachUserToCompany(iCid, uid, cid string) error {
	//FindIntercomUser
	user := &User{
		Id: uid,
	}
	u, err := c.matchUser(user)
	if err != nil {
		return err
	}
	if u == nil {
		c.Log("intercom-user-not-found: " + user.Id)
		return errors.New("intercom-user-not-found")
	}
	c.Log("User Matched::" + u.ID)

	if iCid == "" {
		company, err := c.ic.Companies.FindByCompanyID(cid)
		if err != nil {
			if !isNotFound(err) {
				c.Log("error-find-by-company-id: " + cid)
				c.Log(err.Error())
				return err
			}
			//Lets create
			company, err = c.ic.Companies.Save(&intercom.Company{CompanyID: cid, Name: "Company_" + cid})
			if err != nil {
				c.Log("unable-to-create-new-company: " + cid)
				return errors.New("unable-to-create-new-company")
			}
		}
		if company.ID == "" {
			if err != nil {
				c.Log("unable-to-fetch-company: " + cid)
				return errors.New("unable-to-fetch-company")
			}
		}
		iCid = company.ID
	}

	if u.Companies == nil {
		u.Companies = &intercom.CompanyList{
			Pages:       intercom.PageParams{},
			Companies:   make([]intercom.Company, 0),
			ScrollParam: "",
		}
	}
	for _, co := range u.Companies.Companies {
		if co.CompanyID == cid || co.ID == iCid {
			//Already in
			return nil
		}
	}
	u.Companies.Companies = append(u.Companies.Companies, intercom.Company{CompanyID: cid, ID: iCid})
	_, err = c.ic.Users.Save(u)
	return err
}
