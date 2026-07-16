package events

const (
	BookingCreatedQueue = "booking.created"
	PaymentSuccessQueue = "payment.success"
)

type BookingCreated struct {
	BookingID string `json:"booking_id"`
	EventID   string `json:"event_id"`
	UserID    string `json:"user_id"`
}

type PaymentSuccess struct {
	BookingID string `json:"booking_id"`
	EventID   string `json:"event_id"`
	UserID    string `json:"user_id"`
}
