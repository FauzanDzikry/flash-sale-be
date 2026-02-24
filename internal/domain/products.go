package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;"`
	Name      string     `gorm:"type:varchar(255);not null"`
	Category  string     `gorm:"type:varchar(255);not null"`
	Stock     int        `gorm:"type:int;not null"`
	Price     float64    `gorm:"type:decimal(10,2);not null"`
	Discount  float64    `gorm:"type:decimal(10,2);not null"`
	CreatedAt time.Time  `gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt time.Time  `gorm:"type:timestamp;not null;default:now()"`
	DeletedAt *time.Time `gorm:"type:timestamp;"`
	CreatedBy uuid.UUID  `gorm:"type:uuid;not null"`
}

func (p *Product) TableName() string {
	return "products"
}

func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
