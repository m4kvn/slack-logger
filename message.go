package main

type Message struct {
	Type string `json:"type"`
	User string `json:"user"`
	Text string `json:"text"`
	Ts   string `json:"ts"`
}

type Attachment struct {
	Text           string `json:"text"`
	Fallback       string `json:"fallback"`
	CallbackId     string `json:"callback_id"`
	Color          string `json:"color"`
	AttachmentType string `json:"attachment_type"`
	Pretext        string `json:"pretext"`
}
