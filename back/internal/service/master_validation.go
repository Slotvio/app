package service

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	ErrInvalidInput = errors.New("invalid input")

	profileSlugPattern = regexp.MustCompile(`^[a-z0-9_-]{3,40}$`)
)

const (
	minServiceDurationMinutes = 5
	maxServiceDurationMinutes = 480
	maxSchedulePeriodsPerDay  = 4
)

func invalidField(field, requirement string) error {
	return fmt.Errorf("%w: %s %s", ErrInvalidInput, field, requirement)
}

func validateUserID(userID int64) error {
	if userID <= 0 {
		return invalidField("user_id", "must be positive")
	}
	return nil
}

func validateServiceID(serviceID int64) error {
	if serviceID <= 0 {
		return invalidField("service_id", "must be positive")
	}
	return nil
}

func validateProfileSlug(slug string) error {
	if !profileSlugPattern.MatchString(slug) {
		return invalidField("slug", "must match ^[a-z0-9_-]{3,40}$")
	}
	return nil
}

func validateTimezone(timezone string) error {
	if strings.TrimSpace(timezone) == "" {
		return invalidField("timezone", "is required")
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return invalidField("timezone", "must be a valid IANA timezone")
	}
	return nil
}

func validateCreateProfileInput(userID int64, input CreateProfileInput) error {
	if err := validateUserID(userID); err != nil {
		return err
	}
	if err := validateProfileSlug(input.Slug); err != nil {
		return err
	}
	if strings.TrimSpace(input.Title) == "" {
		return invalidField("title", "is required")
	}
	if strings.TrimSpace(input.Category) == "" {
		return invalidField("category", "is required")
	}
	return validateTimezone(input.Timezone)
}

func validateGetOwnProfileInput(userID int64) error {
	return validateUserID(userID)
}

func validateGetPublicProfileInput(slug string) error {
	return validateProfileSlug(slug)
}

func validateUpdateProfileInput(userID int64, input UpdateProfileInput) error {
	if err := validateUserID(userID); err != nil {
		return err
	}
	if input.Title == nil && input.Category == nil && input.Description == nil && input.Timezone == nil {
		return invalidField("profile", "must contain at least one field to update")
	}
	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return invalidField("title", "cannot be empty")
	}
	if input.Category != nil && strings.TrimSpace(*input.Category) == "" {
		return invalidField("category", "cannot be empty")
	}
	if input.Timezone != nil {
		return validateTimezone(*input.Timezone)
	}
	return nil
}

func validateListServicesInput(userID int64) error {
	return validateUserID(userID)
}

func validateCreateServiceInput(userID int64, input CreateServiceInput) error {
	if err := validateUserID(userID); err != nil {
		return err
	}
	if strings.TrimSpace(input.Name) == "" {
		return invalidField("name", "is required")
	}
	if input.DurationMinutes < minServiceDurationMinutes || input.DurationMinutes > maxServiceDurationMinutes {
		return invalidField("duration_minutes", "must be between 5 and 480")
	}
	if input.Price < 0 {
		return invalidField("price", "must not be negative")
	}
	return nil
}

func validateUpdateServiceInput(userID, serviceID int64, input UpdateServiceInput) error {
	if err := validateUserID(userID); err != nil {
		return err
	}
	if err := validateServiceID(serviceID); err != nil {
		return err
	}
	if input.Name == nil && input.Description == nil && input.DurationMinutes == nil && input.Price == nil && input.IsActive == nil {
		return invalidField("service", "must contain at least one field to update")
	}
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return invalidField("name", "cannot be empty")
	}
	if input.DurationMinutes != nil && (*input.DurationMinutes < minServiceDurationMinutes || *input.DurationMinutes > maxServiceDurationMinutes) {
		return invalidField("duration_minutes", "must be between 5 and 480")
	}
	if input.Price != nil && *input.Price < 0 {
		return invalidField("price", "must not be negative")
	}
	return nil
}

func validateReplaceScheduleInput(userID int64, input ReplaceScheduleInput) error {
	if err := validateUserID(userID); err != nil {
		return err
	}

	periodsByDay := make(map[int8][]clockInterval, 7)
	for _, period := range input.Periods {
		if period.DayOfWeek < 1 || period.DayOfWeek > 7 {
			return invalidField("day_of_week", "must be between 1 and 7 (ISO weekday)")
		}
		if len(periodsByDay[period.DayOfWeek]) >= maxSchedulePeriodsPerDay {
			return invalidField("periods", "must not contain more than 4 intervals per day")
		}

		start, err := parseClockTime("start_time", period.StartTime)
		if err != nil {
			return err
		}
		end, err := parseClockTime("end_time", period.EndTime)
		if err != nil {
			return err
		}
		if start >= end {
			return invalidField("period", "start_time must be earlier than end_time")
		}

		periodsByDay[period.DayOfWeek] = append(periodsByDay[period.DayOfWeek], clockInterval{start: start, end: end})
	}

	for _, intervals := range periodsByDay {
		sort.Slice(intervals, func(i, j int) bool { return intervals[i].start < intervals[j].start })
		for i := 1; i < len(intervals); i++ {
			if intervals[i].start < intervals[i-1].end {
				return invalidField("periods", "must not overlap within a day")
			}
		}
	}

	return nil
}

func validateGetScheduleInput(userID int64) error {
	return validateUserID(userID)
}

type clockInterval struct {
	start time.Duration
	end   time.Duration
}

func parseClockTime(field, value string) (time.Duration, error) {
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return 0, invalidField(field, "must use 24-hour HH:MM format")
	}
	return time.Duration(parsed.Hour())*time.Hour + time.Duration(parsed.Minute())*time.Minute, nil
}
