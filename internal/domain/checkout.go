package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Checkout struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null"`
	ProductID  uuid.UUID  `gorm:"type:uuid;not null"`
	Quantity   int        `gorm:"type:int;not null"`
	Price      float64    `gorm:"type:decimal(10,2);not null"`
	Discount   float64    `gorm:"type:decimal(10,2);not null"`
	TotalPrice float64    `gorm:"type:decimal(10,2);not null"`
	CreatedAt  time.Time  `gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt  time.Time  `gorm:"type:timestamp;not null;default:now()"`
	DeletedAt  *time.Time `gorm:"type:timestamp;"`
}

func (c *Checkout) TableName() string {
	return "checkouts"
}

func (c *Checkout) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
