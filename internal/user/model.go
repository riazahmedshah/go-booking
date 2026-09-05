package user

type User struct {
	ID        string  `json:"id" db:"id"`
	FirstName string  `json:"firstName" db:"first_name"`
	LastName  *string `json:"lastName" db:"last_name"`
	Email     string  `json:"email" db:"email"`
	Role      string  `json:"role" db:"role"`
}

type UserEmail struct {
	Email string `json:"email" db:"email"`
}
