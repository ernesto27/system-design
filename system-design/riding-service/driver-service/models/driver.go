package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Driver struct {
	ID            uint      `json:"id" gorm:"primary_key"`
	Email         string    `json:"email" gorm:"unique;not null"`
	Password      string    `json:"password" gorm:"not null"`
	Name          string    `json:"name"`
	LicenseNumber string    `json:"license_number" gorm:"unique;not null"`
	VehicleInfo   Vehicle   `json:"vehicle" gorm:"foreignKey:DriverID"`
	IsAvailable   bool      `json:"is_available" gorm:"default:false"`
	CurrentLat    float64   `json:"current_lat"`
	CurrentLng    float64   `json:"current_lng"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Vehicle struct {
	ID           uint   `json:"id" gorm:"primary_key"`
	DriverID     uint   `json:"driver_id"`
	Model        string `json:"model"`
	PlateNumber  string `json:"plate_number" gorm:"unique"`
	VehicleType  string `json:"vehicle_type"`
	Manufacturer string `json:"manufacturer"`
	Year         int    `json:"year"`
}

func (d *Driver) BeforeCreate(tx *gorm.DB) error {
	if len(d.Password) > 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(d.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		d.Password = string(hashedPassword)
	}
	return nil
}
