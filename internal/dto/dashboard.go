package dto

import "time"

// DashboardRead is the payload for the common web dashboard. Widgets are
// returned in the stable product order, with unavailable widgets omitted.
type DashboardRead struct {
	Greeting     DashboardGreetingRead `json:"greeting"`
	Widgets      []DashboardWidgetRead `json:"widgets"`
	Empty        bool                  `json:"empty"`
	EmptyMessage string                `json:"empty_message,omitempty"`

	// UnreadNotificationCount backs the mobile dashboard header bell
	// (DR-MOB-005-001-01 Rule 14). Unlike the widgets, this is never omitted:
	// notifications carry no permission gate, so every caller has a count.
	// Named in full because at the dashboard root a bare "unread_count" does
	// not say unread what.
	UnreadNotificationCount int64 `json:"unread_notification_count"`
}

type DashboardGreetingRead struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type DashboardWidgetRead struct {
	ID           string                `json:"id"`
	Title        string                `json:"title"`
	Scope        string                `json:"scope"`
	Metrics      []DashboardMetricRead `json:"metrics"`
	Items        []DashboardItemRead   `json:"items"`
	Actions      []DashboardActionRead `json:"actions"`
	EmptyMessage string                `json:"empty_message,omitempty"`
}

type DashboardMetricRead struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value any    `json:"value"`
}

type DashboardItemRead struct {
	ID          string     `json:"id,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
	URL         string     `json:"url,omitempty"`
}

type DashboardActionRead struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	URL   string `json:"url"`
}
