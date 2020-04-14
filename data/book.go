package data

// Book data model.
type Book struct {
	ID     int    `db:"id,omitempty" json:"id"`
	Title  string `db:"title" json:"title"`
	Author string `db:"author" json:"author"`
}
