package main

type Activity struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	Status    string `json:"status"`
	UserID    int64  `json:"userId"`
	CoverURL  string `json:"coverUrl,omitempty"`
	Summary   string `json:"summary,omitempty"`
	Content   string `json:"content,omitempty"`
	Location  string `json:"location,omitempty"`
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type ActivityRegistration struct {
	ID         int64  `json:"id"`
	ActivityID int64  `json:"activityId"`
	UserID     int64  `json:"userId"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
}

type Service struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Address     string `json:"address,omitempty"`
	Description string `json:"description,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type LostItem struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	ItemType    string `json:"itemType,omitempty"`
	Status      string `json:"status,omitempty"`
	Location    string `json:"location,omitempty"`
	OccurredAt  string `json:"occurredAt,omitempty"`
	Description string `json:"description,omitempty"`
	Contact     string `json:"contact,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type listResponse[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
	User      struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
}
