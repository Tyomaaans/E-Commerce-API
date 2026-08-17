package categories

// ==== Request ====

type AddCategoryParentRequest struct {
	Name  string  `json:"name"  validate:"required,alphaspaceunicode"`
	Slug  string  `json:"-"`
	Image *string `json:"image" validate:"omitempty,url"`
}

type AddCategoryChildRequest struct {
	ID       string  `json:"-"`
	Name     string  `json:"name" validate:"required,alphaspaceunicode"`
	Slug     string  `json:"-"`
	Image    *string `json:"image" validate:"omitempty,url"`
	ParentID string  `json:"-"`
}

type UpdateCategoryRequest struct {
	ID       string  `json:"-"`
	Name     *string `json:"name"  validate:"omitempty,alphaspaceunicode"`
	Slug     *string `json:"-"`
	Image    *string `json:"image" validate:"omitempty,url"`
	ParentID string  `json:"-"`
}

type MoveCategoryParentRequest struct {
    ID       string `json:"-"`
    ParentID string `json:"parent_id"`
}

// ==== Response ====

type CategoryResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	Image    *string `json:"image,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`

	Children []*CategoryResponse `json:"children,omitempty"`
}