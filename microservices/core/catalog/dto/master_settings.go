package dto

//go:generate easyjson $GOFILE

// MasterSettings — настройки расписания мастера.
//
//easyjson:json
type MasterSettings struct {
	MasterID        string `json:"master_id"`
	SlotStepMinutes int    `json:"slot_step_minutes"`
	LeadTimeMinutes int    `json:"lead_time_minutes"`
}

// UpsertMasterSettingsRequest — частичное обновление настроек.
// Поле NULL = не менять.
//
//easyjson:json
type UpsertMasterSettingsRequest struct {
	SlotStepMinutes *int `json:"slot_step_minutes,omitempty"`
	LeadTimeMinutes *int `json:"lead_time_minutes,omitempty"`
}
