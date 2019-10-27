package model

// Cfg interface return a global configuration
type Cfg interface {
	//Cfg return config
	Cfg() Config
}

// Text interface return message text
type Text interface {
	TextInfo() string
}
