package products

// ==== Request ====

type AddProductRequest struct {
	StoreID     string  `json:"-"`
	CategoryID  string  `json:"category_id"  validate:"required,uuid"`
	Name        string  `json:"name"         validate:"required"`
	Slug        string  `json:"-"`
	Description string  `json:"description"  validate:"required"`
	Price       float64 `json:"price"        validate:"required,gt=0"`
	Stock       int     `json:"stock"        validate:"min=0"`
	WeightGram  int     `json:"weight_gram"  validate:"required,gt=0"`
}

type UpdateProductRequest struct {
	ID          string   `json:"-"`
	StoreID     string   `json:"-"`
	CategoryID  *string  `json:"category_id"  validate:"omitempty,uuid"`
	Name        *string  `json:"name"         validate:"omitempty"`
	Slug        *string  `json:"-"`
	Description *string  `json:"description"  validate:"omitempty"`
	Price       *float64 `json:"price"        validate:"omitempty,gt=0"`
	Stock       *int     `json:"stock"        validate:"omitempty,min=0"`
	WeightGram  *int     `json:"weight_gram"  validate:"omitempty,gt=0"`
}

type AddProductImageRequest struct {
	ProductID   string `json:"-"`
	StoreID     string `json:"-"`
	ImageURL    string `json:"image_url" validate:"required,url"`
	IsPrimary   bool   `json:"is_primary"`
}

type UpdateProductImageRequest struct {
	ID          string  `json:"-"`
	ProductID   string  `json:"-"`
	StoreID     string  `json:"-"`
	ImageURL    *string `json:"image_url" validate:"omitempty,url"`
	IsPrimary   *bool   `json:"is_primary" validate:"omitempty"`
}

type AddProductReviewRequest struct {
	UserID      string  `json:"-"`
	ProductID   string  `json:"-"`
	Rating      int     `json:"rating"  validate:"required,min=1,max=5"`
	Comment     *string `json:"comment" validate:"omitempty"`
}

type UpdateProductReviewRequest struct {
	ID          string  `json:"-"`
	UserID      string  `json:"-"`
	ProductID   string  `json:"-"`
	Rating      *int    `json:"rating"  validate:"omitempty,min=1,max=5"`
	Comment     *string `json:"comment" validate:"omitempty"`
}

// ==== Response ====

type ProductImageResponse struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	ImageURL  string `json:"image_url"`
	IsPrimary bool   `json:"is_primary"`
}

type ProductReviewResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	ProductID string  `json:"product_id"`
	Rating    int     `json:"rating"`
	Comment   *string `json:"comment,omitempty"`
}

type ProductResponse struct {
	ID            string  `json:"id"`
	StoreID       string  `json:"store_id"`
	CategoryID    string  `json:"category_id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	Description   string  `json:"description"`
	Price         float64 `json:"price"`
	Stock         int     `json:"stock"`
	ReservedStock int     `json:"reserved_stock"`
	WeightGram    int     `json:"weight_gram"`
	IsActive      bool    `json:"is_active"`

	Images  []ProductImageResponse  `json:"images,omitempty"`
	Reviews []ProductReviewResponse `json:"reviews,omitempty"`
}

/*
================================================================================
                                1. PRODUCT ENDPOINTS
================================================================================

// -----------------------------------------------------------------------------
// GET /api/v1/stores/{store_id}/products
// -----------------------------------------------------------------------------
// Fungsi       : Mengambil daftar produk berdasarkan toko (pagination & filter).
// Path Params  : store_id (string)
// Query Params : page (int), limit (int), category_id (string), search (string)
// Response     : Array of ProductEntity

// -----------------------------------------------------------------------------
// GET /api/v1/stores/{store_id}/products/{id}
// -----------------------------------------------------------------------------
// Fungsi       : Mengambil detail satu produk spesifik.
// Path Params  : store_id (string), id (string - Product ID)
// Response     : ProductEntity (termasuk Images & Reviews jika diperlukan)

// -----------------------------------------------------------------------------
// POST /api/v1/stores/{store_id}/products
// -----------------------------------------------------------------------------
// Fungsi       : Membuat produk baru di toko tertentu.
// Path Params  : store_id (string)
// Body Payload : ProductEntity (tanpa ID, biasanya ID digenerate oleh backend)
// Response     : ProductEntity (terbuat)

// -----------------------------------------------------------------------------
// PATCH /api/v1/stores/{store_id}/products/{id}
// -----------------------------------------------------------------------------
// Fungsi       : Mengupdate data produk secara parsial (hanya field yang dikirim).
// Path Params  : store_id (string), id (string - Product ID)
// Body Payload : UpdateProductEntity
// Response     : ProductEntity (ter-update)

// -----------------------------------------------------------------------------
// DELETE /api/v1/stores/{store_id}/products/{id}
// -----------------------------------------------------------------------------
// Fungsi       : Menghapus produk tertentu dari toko.
// Path Params  : store_id (string), id (string - Product ID)
// Response     : Success Message / Status 204 No Content


================================================================================
                             2. PRODUCT IMAGES ENDPOINTS
================================================================================

// -----------------------------------------------------------------------------
// POST /api/v1/products/{product_id}/images
// -----------------------------------------------------------------------------
// Fungsi       : Menambahkan/upload gambar baru ke produk tertentu.
// Path Params  : product_id (string)
// Body Payload : ProductImageEntity
// Response     : ProductImageEntity

// -----------------------------------------------------------------------------
// PATCH /api/v1/products/{product_id}/images/{image_id}
// -----------------------------------------------------------------------------
// Fungsi       : Mengupdate atribut gambar (contoh: mengubah status IsPrimary).
// Path Params  : product_id (string), image_id (string)
// Body Payload : UpdateProductImageEntity
// Response     : ProductImageEntity

// -----------------------------------------------------------------------------
// DELETE /api/v1/products/{product_id}/images/{image_id}
// -----------------------------------------------------------------------------
// Fungsi       : Menghapus gambar tertentu dari produk.
// Path Params  : product_id (string), image_id (string)
// Response     : Success Message / Status 204 No Content


================================================================================
                             3. PRODUCT REVIEWS ENDPOINTS
================================================================================

// -----------------------------------------------------------------------------
// GET /api/v1/products/{product_id}/reviews
// -----------------------------------------------------------------------------
// Fungsi       : Mengambil daftar ulasan untuk produk tertentu.
// Path Params  : product_id (string)
// Query Params : page (int), limit (int), rating (int)
// Response     : Array of ProductReviewEntity

// -----------------------------------------------------------------------------
// POST /api/v1/products/{product_id}/reviews
// -----------------------------------------------------------------------------
// Fungsi       : Menambahkan ulasan baru pada produk.
// Path Params  : product_id (string)
// Body Payload : ProductReviewEntity
// Response     : ProductReviewEntity

// -----------------------------------------------------------------------------
// PATCH /api/v1/products/{product_id}/reviews/{review_id}
// -----------------------------------------------------------------------------
// Fungsi       : Mengedit ulasan yang sudah dibuat user.
// Path Params  : product_id (string), review_id (string)
// Body Payload : UpdateReviewEntity
// Response     : ProductReviewEntity

// -----------------------------------------------------------------------------
// DELETE /api/v1/products/{product_id}/reviews/{review_id}
// -----------------------------------------------------------------------------
// Fungsi       : Menghapus ulasan tertentu.
// Path Params  : product_id (string), review_id (string)
// Response     : Success Message / Status 204 No Content
*/