package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/go-playground/validator.v9"
)

const (
	ErrorFirebaseEmailMissing         = "no firebase email available"
	ErrorBlockedEmailDomain           = "blocked-email-domain"
	ErrorNotExistentTranslation       = "not-existent-translation"
	ErrorPublishedContent             = "cannot-remove-published-content"
	ErrorResourceNotFound             = "resource-not-found"
	ErrorBookNotFound                 = "book-not-found"
	ErrorUnmatchedUser                = "unmatched-user"
	ErrorEmptyArray                   = "empty-array-provided"
	ErrorEmptyBook                    = "empty-book-provided"
	ErrorAnalyticsPlatformUnsupported = "unsupported-platform"
	ErrorAnalyticsEventUnsupported    = "unsupported-event_type"
	ErrorInvalidObjectId              = "invalid-object-id"
	ErrorSubSubscriptionNotFound      = "subscription-not-found"
	ErrorSubAlreadyExists             = "subscription-already-in-database"
	ErrorSubOwnedByOtherUser          = "subscription-owned-by-different-user"
	ErrorUserNotFound                 = "user-not-found"
	ErrorCMSRightsNeeded              = "request-need-cms-access"
	ErrorCMSAdminRightsNeeded         = "request-need-cms-admin-access"
	ErrorBookCurrentNotFilled         = "book-current-not-filled"
	ErrorIntercomEventUnsupported     = "unsupported-intercom-event_type"
	ErrorRequiresSubscription         = "requires_subscription"
	ErrorPublishedContentWithoutDate  = "published-contents-must-have-a-publish-date"
	ErrorSlugPresent                  = "slug-not-unique"
	ErrorConcurrency                  = "record-was-already-changed"
	ErrorSubscriptionItemPlanNotFound = "subscription-item-plan-not-found"
)

type NotFound struct {
	Err error
}

func (e NotFound) Error() string {
	return ErrorResourceNotFound
}

func (e NotFound) Code() int {
	return 404
}

type Forbidden struct {
	Err error
}

func (e Forbidden) Error() string {
	if e.Err == nil {
		return "no rights to access the content"
	}

	return e.Err.Error()
}

func (e Forbidden) Code() int {
	return 403
}

type InvalidRequestError struct {
	Err error
}

func (e InvalidRequestError) Error() string {
	if e.Err == nil {
		return "invalid-request"
	}

	return e.Err.Error()
}

func (e InvalidRequestError) Code() int {
	return http.StatusUnprocessableEntity
}

type PreconditionFailedError struct {
	Err error
}

func (e PreconditionFailedError) Error() string {
	if e.Err == nil {
		return "precondition-failed"
	}

	return e.Err.Error()
}

func (e PreconditionFailedError) Code() int {
	return http.StatusPreconditionFailed
}

type ValidationError struct {
	Children validator.ValidationErrors
}

func (v ValidationError) Error() string {
	var b strings.Builder
	for _, err := range v.Children {
		template := fmt.Sprintf("%s=%s|", err.StructField(), err.Tag())
		b.WriteString(template)
	}

	return b.String()
}

func (v ValidationError) Code() int {
	return http.StatusUnprocessableEntity
}

func (v ValidationError) MarshalJSON() ([]byte, error) {
	content := make(map[string]string)
	for _, err := range v.Children {
		content[err.Field()] = fmt.Sprintf("failed '%s' validation", err.Tag())
	}

	return json.Marshal(content)
}

func PrintVarDump(title, i interface{}) {
	fmt.Printf("\n------\n%s \n %s \n", title, VarDump(i))
}

func VarDump(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", " ")
	return string(s)
}
