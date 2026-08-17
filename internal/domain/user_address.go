package domain

type UserAddressEntity struct {
    ID           string
    UserID       string
    Label        string
    Recipient    string
    Phone        string
    AddressLine1 string
    AddressLine2 *string
    City         string
    State        string
    PostalCode   string
    Country      string
    IsPrimary    *bool
}

type UpdateUserAddressEntity struct {
    ID           string
    UserID       string
    Label        *string
    Recipient    *string
    Phone        *string
    AddressLine1 *string
    AddressLine2 *string
    City         *string
    State        *string
    PostalCode   *string
    Country      *string
    IsPrimary    *bool
}