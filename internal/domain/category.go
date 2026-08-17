package domain

type CategoryEntity struct {
	ID        string
	Name      string
	Slug      string
	Image     *string
	ParentID  *string

	Children []*CategoryEntity
	Products []ProductEntity
}

type UpdateCategoryEntity struct {
	ID        string
	Name      *string
	Slug      *string
	Image     *string
	ParentID  *string
}