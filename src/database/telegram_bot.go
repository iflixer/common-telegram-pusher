package database

type TelegramBot struct {
	ID           int
	Name         string
	Type         string
	BotURL       string
	AppURL       string
	Description  string
	Token        string
	GaTrackingID string
	GaSecret     string
	SearchURL    string
	UpdatedAt    string
	Published    bool
}

func (b *TelegramBot) TableName() string {
	return "telegram_bot"
}

func (c *TelegramBot) Load(dbService *Service, id int) (err error) {
	return dbService.DB.Where("id=?", id).Limit(1).First(&c).Error
}
