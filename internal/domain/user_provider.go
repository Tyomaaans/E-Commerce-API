package domain

type Provider string

const (
    ProviderCredentials Provider = "credentials"
    ProviderGoogle      Provider = "google"
    ProviderFacebook    Provider = "facebook"
)

type UserProviderEntity struct {
    ID             string
    UserID         string
    Provider       Provider
    ProviderUserID *string
}

type CreateUserProviderEntity struct {
    UserID         string
    Provider       Provider
    ProviderUserID *string
    Password       *string
}