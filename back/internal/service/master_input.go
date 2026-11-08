package service

type CreateProfileInput struct {
	Slug        string
	Title       string
	Category    string
	Description string
	Timezone    string
}

type UpdateProfileInput struct {
	Title       *string
	Category    *string
	Description *string
	Timezone    *string
}

type CreateServiceInput struct {
	Name            string
	Description     string
	DurationMinutes int
	Price           int
}

type UpdateServiceInput struct {
	Name            *string
	Description     *string
	DurationMinutes *int
	Price           *int
	IsActive        *bool
}

type SchedulePeriodInput struct {
	DayOfWeek int8
	StartTime string
	EndTime   string
}

type ReplaceScheduleInput struct {
	Periods []SchedulePeriodInput
}
