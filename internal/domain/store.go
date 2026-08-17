package domain

type StoreEntity struct {
	ID          string
	UserID      string
	Name        string
	Slug        string
	Description *string
	Logo        *string
	IsActive    bool

	User        *UserEntity
	Products    []ProductEntity
}

type UpdateStoreEntity struct {
	ID          string
	UserID      string
	Name        *string
	Slug        *string
	Description *string
	Logo        *string
}