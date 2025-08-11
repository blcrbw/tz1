package subscription

type Subscription struct {
	ID          string `json:"id"`
	ServiceName string `json:"service_name"`
	Price       uint   `json:"price"`
	User        string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
}
