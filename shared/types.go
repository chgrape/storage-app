package shared

type Metadata struct {
	Filename string
	Size     int64
	MimeType string
}

type Config struct {
	Host string
	User string
	Pass string
	Port string
	DB   string
}
