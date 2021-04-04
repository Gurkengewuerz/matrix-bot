package types

type Message struct {
	Message         string
	Response        string
	Sender          string
	SenderLocalPart string
	RoomID          string
	EventType       string
	Canceled        bool
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

type HTTPRequest struct {
	Url        string
	Body        map[string]string
	Headers     map[string]string
	Method      string
}

type HTTPResponse struct {
	StatusCode  int
	Headers     map[string]string
	Body        string
}
