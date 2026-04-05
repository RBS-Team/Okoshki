package dto

//go:generate easyjson $GOFILE

//easyjson:json
type WorkingHours struct {
	ID        string  `json:"id"`
	MasterID  string  `json:"master_id"`
	DayOfWeek int     `json:"day_of_week"`
	StartTime *string `json:"start_time,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`
	IsDayOff  bool    `json:"is_day_off"`
}

//easyjson:json
type WorkingHoursDayRequest struct {
	DayOfWeek int     `json:"day_of_week"`
	StartTime *string `json:"start_time,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`
	IsDayOff  bool    `json:"is_day_off"`
}

//easyjson:json
type UpdateWorkingHoursBulkRequest struct {
	Days []WorkingHoursDayRequest `json:"days"`
}

//easyjson:json
type ScheduleException struct {
	ID            string  `json:"id"`
	MasterID      string  `json:"master_id"`
	ExceptionDate string  `json:"exception_date"`
	StartTime     *string `json:"start_time,omitempty"`
	EndTime       *string `json:"end_time,omitempty"`
	IsWorking     bool    `json:"is_working"`
}

//easyjson:json
type CreateScheduleExceptionRequest struct {
	ExceptionDate string  `json:"exception_date"`
	StartTime     *string `json:"start_time,omitempty"`
	EndTime       *string `json:"end_time,omitempty"`
	IsWorking     bool    `json:"is_working"`
}

//easyjson:json
type UpdateScheduleExceptionRequest struct {
	StartTime *string `json:"start_time,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`
	IsWorking *bool   `json:"is_working,omitempty"`
}
