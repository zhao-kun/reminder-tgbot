package model

// Cfg interface return a global configuration
type Cfg interface {
	Cfg() Config
}

// Text interface return message text
type Text interface {
	TextInfo() string
}
