package utils

import "net/http"

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
