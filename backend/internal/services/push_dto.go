package services

// PushSubscribeInput mirrors the browser PushSubscription JSON shape.
type PushSubscribeInput struct {
	Endpoint string `json:"endpoint" validate:"required,url"`
	Keys     struct {
		P256dh string `json:"p256dh" validate:"required"`
		Auth   string `json:"auth" validate:"required"`
	} `json:"keys" validate:"required"`
	UserAgent string `json:"user_agent" validate:"omitempty,max=255"`
}

// PushUnsubscribeInput identifies a subscription to remove.
type PushUnsubscribeInput struct {
	Endpoint string `json:"endpoint" validate:"required,url"`
}
