package store

type Agreement struct {
	ID          uint   `gorm:"primaryKey"`
	Version     string `gorm:"index"`
	EULA        string
	PrivacyHTML string
}
