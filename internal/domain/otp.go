package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OTP struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	Email     string    `gorm:"type:varchar(50);not null"`
	OTPCode   string    `gorm:"type:varchar(10);not null"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null"`
	Used      bool      `gorm:"type:boolean;not null;default:false"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
}

func (OTP) TableName() string {
	return "otps"
}

func (o *OTP) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
