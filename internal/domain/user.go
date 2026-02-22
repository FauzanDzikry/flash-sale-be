package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;"`
	Email         string     `gorm:"type:varchar(50);unique;not null"`
	Password      string     `gorm:"column:password_hash;type:varchar(255);not null"`
	Name          string     `gorm:"type:varchar(100);"`
	CreatedAt     time.Time  `gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt     time.Time  `gorm:"type:timestamp;not null;default:now()"`
	DeactivatedAt *time.Time `gorm:"type:timestamp;"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
