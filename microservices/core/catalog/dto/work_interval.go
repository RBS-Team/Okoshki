package dto

//go:generate easyjson $GOFILE

// WorkInterval — рабочий интервал мастера на конкретную дату.
// Date в формате "2006-01-02"; StartTime/EndTime в формате "15:04".
//
//easyjson:json
type WorkInterval struct {
	ID        string `json:"id"`
	MasterID  string `json:"master_id"`
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// IntervalInput — описание одного интервала без идентификатора.
//
//easyjson:json
type IntervalInput struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// CreateWorkIntervalRequest — создать один интервал на конкретную дату.
//
//easyjson:json
type CreateWorkIntervalRequest struct {
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// ReplaceWorkIntervalsForDateRequest — атомарно заменить все интервалы на дату.
// Пустой список = очистить день (мастер не работает).
//
//easyjson:json
type ReplaceWorkIntervalsForDateRequest struct {
	Date      string          `json:"date"`
	Intervals []IntervalInput `json:"intervals"`
}

// WorkIntervalList — список интервалов в диапазоне дат.
//
//easyjson:json
type WorkIntervalList struct {
	Intervals []WorkInterval `json:"intervals"`
}
