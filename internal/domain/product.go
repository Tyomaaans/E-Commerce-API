package domain

type ProductEntity struct {
	ID            string
	StoreID       string
	CategoryID    string
	Name          string
	Slug          string
	Description   string
	Price         float64
	Stock         int
	ReservedStock int
	WeightGram    int
	IsActive      bool
	Images        []ProductImageEntity
	Reviews		  []ProductReviewEntity 
}

type ProductImageEntity struct {
	ID        string
	ProductID string
	ImageURL  string
	IsPrimary bool
}

type ProductReviewEntity struct {
	ID        string
	UserID    string
	ProductID string
	Rating    int
	Comment   *string
}

type UpdateProductEntity struct {
	ID            string
	StoreID       string
	CategoryID    *string
	Name          *string
	Slug          *string
	Description   *string
	Price         *float64
	Stock         *int
	ReservedStock *int
	WeightGram    *int
}

type UpdateProductImageEntity struct {
	ID        string
	ProductID string
	ImageURL  *string
	IsPrimary *bool
}

type UpdateReviewEntity struct {
	ID        string
	UserID    string
	ProductID string
	Rating    *int
	Comment   *string
}