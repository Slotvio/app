package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aleksandrboro/slotvio/internal/domain"
)

var (
	ErrInvalidBookingInput      = errors.New("invalid booking input")
	ErrBookableResourceNotFound = errors.New("master or active service not found")
	ErrBookingSlotUnavailable   = errors.New("booking slot is unavailable")
	ErrBookingRepositoryMissing = errors.New("booking repository is not configured")
)

const (
	bookingSlotStep       = 30 * time.Minute
	bookingMinLeadTime    = 2 * time.Hour
	bookingMaxHorizonDays = 60
	maxClientCommentChars = 1000
)

type BookingRepository interface {
	GetMasterBySlug(ctx context.Context, slug string) (domain.Master, error)
	GetActiveServiceByID(ctx context.Context, masterID, serviceID int64) (domain.Service, error)
	ListSchedulePeriodsByDay(ctx context.Context, masterID int64, dayOfWeek int8) ([]domain.SchedulePeriod, error)
	ListActiveBookingsOverlapping(ctx context.Context, masterID int64, startAt, endAt time.Time) ([]domain.Booking, error)
	CreateBooking(ctx context.Context, booking domain.Booking) (domain.Booking, error)
}

type CreateBookingInput struct {
	MasterSlug    string
	ServiceID     int64
	StartAt       time.Time
	ClientComment string
}

type AvailableSlot struct {
	StartAt   time.Time
	LocalTime string
}

type AvailableSlots struct {
	Date     string
	Timezone string
	Slots    []AvailableSlot
}

type bookingService struct {
	repo BookingRepository
	now  func() time.Time
}

func NewBookingService(repo BookingRepository) (*bookingService, error) {
	if repo == nil {
		return nil, ErrBookingRepositoryMissing
	}
	return &bookingService{repo: repo, now: time.Now}, nil
}

func (s *bookingService) GetAvailableSlots(
	ctx context.Context,
	slug string,
	serviceID int64,
	dateValue string,
) (AvailableSlots, error) {
	master, service, location, err := s.loadBookable(ctx, slug, serviceID)
	if err != nil {
		return AvailableSlots{}, err
	}

	parsedDate, err := time.Parse("2006-01-02", dateValue)
	if err != nil || parsedDate.Format("2006-01-02") != dateValue {
		return AvailableSlots{}, fmt.Errorf("%w: date must use YYYY-MM-DD", ErrInvalidBookingInput)
	}
	date := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, location)
	now := s.now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	lastBookableDay := today.AddDate(0, 0, bookingMaxHorizonDays)
	if date.Before(today) || date.After(lastBookableDay) {
		return AvailableSlots{}, fmt.Errorf("%w: date must be within the next %d days", ErrInvalidBookingInput, bookingMaxHorizonDays)
	}

	dayOfWeek := isoWeekday(date)
	periods, err := s.repo.ListSchedulePeriodsByDay(ctx, master.ID, dayOfWeek)
	if err != nil {
		return AvailableSlots{}, fmt.Errorf("list schedule periods: %w", err)
	}

	dayStart := date
	dayEnd := date.AddDate(0, 0, 1)
	bookings, err := s.repo.ListActiveBookingsOverlapping(ctx, master.ID, dayStart.UTC(), dayEnd.UTC())
	if err != nil {
		return AvailableSlots{}, fmt.Errorf("list overlapping bookings: %w", err)
	}

	result := AvailableSlots{
		Date:     dateValue,
		Timezone: master.Timezone,
		Slots:    make([]AvailableSlot, 0),
	}
	earliest := s.now().Add(bookingMinLeadTime)
	latest := now.AddDate(0, 0, bookingMaxHorizonDays)
	duration := time.Duration(service.DurationMinutes) * time.Minute

	for _, period := range periods {
		periodStartMinute := clockMinute(period.StartTime)
		periodEndMinute := clockMinute(period.EndTime)
		periodStart := localClock(date, periodStartMinute, location)
		periodEnd := localClock(date, periodEndMinute, location)

		for minute := periodStartMinute; minute+service.DurationMinutes <= periodEndMinute; minute += int(bookingSlotStep / time.Minute) {
			candidate := localClock(date, minute, location)
			// Skip wall-clock times that do not exist during a DST transition.
			if candidate.Hour()*60+candidate.Minute() != minute {
				continue
			}
			candidateEnd := candidate.Add(duration)
			if candidate.Before(periodStart) || candidateEnd.After(periodEnd) {
				continue
			}
			if candidate.Before(earliest) || candidate.After(latest) {
				continue
			}
			if hasBookingOverlap(candidate, candidateEnd, bookings) {
				continue
			}

			result.Slots = append(result.Slots, AvailableSlot{
				StartAt:   candidate.UTC(),
				LocalTime: candidate.Format("15:04"),
			})
		}
	}

	sort.Slice(result.Slots, func(i, j int) bool {
		return result.Slots[i].StartAt.Before(result.Slots[j].StartAt)
	})
	return result, nil
}

func (s *bookingService) CreateBooking(
	ctx context.Context,
	clientID int64,
	input CreateBookingInput,
) (*domain.Booking, error) {
	if clientID <= 0 || input.ServiceID <= 0 || !profileSlugPattern.MatchString(input.MasterSlug) || input.StartAt.IsZero() {
		return nil, ErrInvalidBookingInput
	}
	if input.StartAt.Second() != 0 || input.StartAt.Nanosecond() != 0 {
		return nil, fmt.Errorf("%w: start_at must be aligned to a whole minute", ErrInvalidBookingInput)
	}
	input.ClientComment = strings.TrimSpace(input.ClientComment)
	if len([]rune(input.ClientComment)) > maxClientCommentChars {
		return nil, fmt.Errorf("%w: client_comment is too long", ErrInvalidBookingInput)
	}

	master, service, location, err := s.loadBookable(ctx, input.MasterSlug, input.ServiceID)
	if err != nil {
		return nil, err
	}

	startAt := input.StartAt.UTC()
	localStart := startAt.In(location)
	now := s.now()
	if startAt.Before(now.Add(bookingMinLeadTime)) || startAt.After(now.In(location).AddDate(0, 0, bookingMaxHorizonDays)) {
		return nil, fmt.Errorf("%w: start_at must be at least two hours ahead and no more than %d days ahead", ErrInvalidBookingInput, bookingMaxHorizonDays)
	}

	day := time.Date(localStart.Year(), localStart.Month(), localStart.Day(), 0, 0, 0, 0, location)
	periods, err := s.repo.ListSchedulePeriodsByDay(ctx, master.ID, isoWeekday(day))
	if err != nil {
		return nil, fmt.Errorf("list schedule periods: %w", err)
	}

	duration := time.Duration(service.DurationMinutes) * time.Minute
	endAt := startAt.Add(duration)
	if !matchesScheduleSlot(localStart, endAt.In(location), day, periods, service.DurationMinutes) {
		return nil, fmt.Errorf("%w: requested time is not a valid working slot", ErrInvalidBookingInput)
	}

	dayStart := day
	dayEnd := day.AddDate(0, 0, 1)
	bookings, err := s.repo.ListActiveBookingsOverlapping(ctx, master.ID, dayStart.UTC(), dayEnd.UTC())
	if err != nil {
		return nil, fmt.Errorf("check booking availability: %w", err)
	}
	if hasBookingOverlap(startAt, endAt, bookings) {
		return nil, ErrBookingSlotUnavailable
	}

	createdAt := now.UTC()
	booking := domain.Booking{
		MasterID:                master.ID,
		ClientID:                clientID,
		ServiceID:               service.ID,
		StartAt:                 startAt,
		EndAt:                   endAt,
		Status:                  "booked",
		ClientComment:           input.ClientComment,
		ServiceNameSnapshot:     service.Name,
		DurationMinutesSnapshot: service.DurationMinutes,
		PriceSnapshot:           service.Price,
		CreatedAt:               createdAt,
		UpdatedAt:               createdAt,
	}

	created, err := s.repo.CreateBooking(ctx, booking)
	if err != nil {
		if errors.Is(err, domain.ErrBookingConflict) {
			return nil, ErrBookingSlotUnavailable
		}
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrBookableResourceNotFound
		}
		return nil, fmt.Errorf("create booking: %w", err)
	}
	return &created, nil
}

func (s *bookingService) loadBookable(
	ctx context.Context,
	slug string,
	serviceID int64,
) (domain.Master, domain.Service, *time.Location, error) {
	if !profileSlugPattern.MatchString(slug) || serviceID <= 0 {
		return domain.Master{}, domain.Service{}, nil, ErrInvalidBookingInput
	}

	master, err := s.repo.GetMasterBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Master{}, domain.Service{}, nil, ErrBookableResourceNotFound
		}
		return domain.Master{}, domain.Service{}, nil, fmt.Errorf("get master by slug: %w", err)
	}
	service, err := s.repo.GetActiveServiceByID(ctx, master.ID, serviceID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Master{}, domain.Service{}, nil, ErrBookableResourceNotFound
		}
		return domain.Master{}, domain.Service{}, nil, fmt.Errorf("get active service: %w", err)
	}
	location, err := time.LoadLocation(master.Timezone)
	if err != nil {
		return domain.Master{}, domain.Service{}, nil, fmt.Errorf("load master timezone: %w", err)
	}
	return master, service, location, nil
}

func hasBookingOverlap(startAt, endAt time.Time, bookings []domain.Booking) bool {
	for _, booking := range bookings {
		if startAt.Before(booking.EndAt) && endAt.After(booking.StartAt) {
			return true
		}
	}
	return false
}

func isoWeekday(date time.Time) int8 {
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return int8(weekday)
}

func clockMinute(value time.Time) int {
	return value.Hour()*60 + value.Minute()
}

func localClock(date time.Time, minute int, location *time.Location) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		minute/60, minute%60, 0, 0, location,
	)
}

func matchesScheduleSlot(
	localStart time.Time,
	localEnd time.Time,
	date time.Time,
	periods []domain.SchedulePeriod,
	durationMinutes int,
) bool {
	startMinute := localStart.Hour()*60 + localStart.Minute()
	endMinute := localEnd.Hour()*60 + localEnd.Minute()
	for _, period := range periods {
		periodStartMinute := clockMinute(period.StartTime)
		periodEndMinute := clockMinute(period.EndTime)
		if startMinute < periodStartMinute || startMinute >= periodEndMinute {
			continue
		}
		if (startMinute-periodStartMinute)%int(bookingSlotStep/time.Minute) != 0 {
			continue
		}
		periodStart := localClock(date, periodStartMinute, localStart.Location())
		periodEnd := localClock(date, periodEndMinute, localStart.Location())
		if localStart.Before(periodStart) || localEnd.After(periodEnd) {
			continue
		}
		if endMinute-startMinute != durationMinutes {
			continue
		}
		return true
	}
	return false
}
