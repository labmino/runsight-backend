package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FullName     string    `json:"full_name" gorm:"type:varchar(100);not null" validate:"required,min=2,max=100"`
	Email        string    `json:"email" gorm:"type:varchar(100);uniqueIndex;not null" validate:"required,email,max=100"`
	Phone        string    `json:"phone,omitempty" gorm:"type:varchar(20)" validate:"omitempty,min=10,max=20"`
	PasswordHash string    `json:"-" gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	Runs []Run `json:"runs,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type UserRegisterRequest struct {
	FullName        string `json:"full_name" validate:"required,min=2,max=100"`
	Email           string `json:"email" validate:"required,email,max=100"`
	Phone           string `json:"phone,omitempty" validate:"omitempty,min=10,max=20"`
	Password        string `json:"password" validate:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserUpdateRequest struct {
	FullName string `json:"full_name,omitempty" validate:"omitempty,min=2,max=100"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,min=10,max=20"`
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}