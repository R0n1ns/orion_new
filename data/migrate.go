package data

func Migrate() {

	DB.AutoMigrate(&User{}, &Message{}, &Channel{}, &Status{})
}
