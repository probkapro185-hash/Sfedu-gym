package domain

import "time"

type Role string

const (
	RoleClient  Role = "client"
	RoleManager Role = "manager"
	RoleAdmin   Role = "admin"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

// User — основная сущность пользователя системы
type User struct {
	ID           int64      `json:"id"`
	FullName     string     `json:"full_name"`
	Phone        string     `json:"phone"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         Role       `json:"role"`
	Gender       Gender     `json:"gender"`
	Balance      float64    `json:"balance"`
	Visits       int        `json:"visits"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastVisitAt  *time.Time `json:"last_visit_at,omitempty"`
}

// ApplicationRequest — заявка от клиента на регистрацию (до принятия менеджером/админом)
type ApplicationRequest struct {
	ID        int64     `json:"id"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Status    string    `json:"status"` // pending, approved, rejected
	CreatedAt time.Time `json:"created_at"`
}

// DTO для создания заявки (публичная форма)
type CreateApplicationInput struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

// DTO для входа в систему
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// DTO для обновления профиля пользователя
type UpdateUserInput struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Gender   Gender `json:"gender"`
}

// DTO для смены пароля
type ChangePasswordInput struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// DTO создания пользователя администратором/менеджером
type CreateUserInput struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     Role   `json:"role"`
	Gender   Gender `json:"gender"`
}
