package dto

// WorkdayQuery is the GET /workdays query string.
type WorkdayQuery struct {
	Year int `form:"year" binding:"required,min=2000,max=2100"`
}

// MonthWorkdaysRead is one row in the monthly workday table.
type MonthWorkdaysRead struct {
	Month     int    `json:"month"`
	MonthName string `json:"month_name"`
	TotalDays int    `json:"total_days"`
	Weekends  int    `json:"weekends"`
	Holidays  int    `json:"holidays"`
	Workdays  int    `json:"workdays"`
}

// WorkdayTotalRead is the summed totals row.
type WorkdayTotalRead struct {
	TotalDays int `json:"total_days"`
	Weekends  int `json:"weekends"`
	Holidays  int `json:"holidays"`
	Workdays  int `json:"workdays"`
}

// WorkdayYearRead is the top-level response payload.
type WorkdayYearRead struct {
	Year   int                 `json:"year"`
	Months []MonthWorkdaysRead `json:"months"`
	Total  WorkdayTotalRead    `json:"total"`
}
