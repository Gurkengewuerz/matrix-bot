package types

type Message struct {
	Message   string
	Response  string
	RoomID    string
	EventType string
	Canceled  bool
}

type HTTPCall struct {
	Path        string
	MatchedPath string
	StatusCode  int
	Body        interface{}
	ContentType string
	Response    string
	Params      map[string]interface{}
	Headers     map[string]string
}
