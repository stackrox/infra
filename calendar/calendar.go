// Package calendar enables the fetching of Google Calendar events used for
// scheduling and launching customer demos.
package calendar

import (
	"context"
	"time"

	"google.golang.org/api/calendar/v3"
)

// calendarCheckTimeout is how long to wait for the Google Calendar API to
// return before timing out.
const calendarCheckTimeout = 10 * time.Second

// Event represents a single calendar event.
type Event struct {
	// ID is a unique identifier for this specific event at at a specific time.
	ID string

	// Email is the email address for the creator of this event.
	Email string

	// Start is the starting time of the event.
	Start time.Time

	// End is the ending time of the event.
	End time.Time

	// Title is the human-readable title of the event.
	Title string

	// Link is a link back to the original Google Calendar event.
	Link string
}

// EventSource represents a type that can fetch calendar events.
type EventSource interface {
	// Events returns a (time-limited) list of events.
	Events() ([]Event, error)
}

type googleCalendar struct {
	window     time.Duration
	calendarID string
	service    *calendar.EventsService
}

// NewGoogleCalendar creates a new Google Calendar connector for fetching events.
func NewGoogleCalendar(calendarID string, window time.Duration) (EventSource, error) {
	service, err := calendar.NewService(context.Background())
	if err != nil {
		return nil, err
	}

	return googleCalendar{
		window:     window,
		calendarID: calendarID,
		service:    service.Events,
	}, nil
}

func (gcal googleCalendar) Events() ([]Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), calendarCheckTimeout)
	defer cancel()

	calEvents, err := gcal.service.List(gcal.calendarID).
		TimeMin(time.Now().Format(time.RFC3339)).
		TimeMax(time.Now().Add(gcal.window).Format(time.RFC3339)).
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}

	events := make([]Event, len(calEvents.Items))
	for index, calEvent := range calEvents.Items {
		end, err := time.Parse(time.RFC3339, calEvent.End.DateTime)
		if err != nil {
			return nil, err
		}

		start, err := time.Parse(time.RFC3339, calEvent.Start.DateTime)
		if err != nil {
			return nil, err
		}

		event := Event{
			ID:    calEvent.Id + "@" + calEvent.Start.DateTime,
			Email: calEvent.Creator.Email,
			Start: start,
			End:   end,
			Title: calEvent.Summary,
			Link:  calEvent.HtmlLink,
		}
		events[index] = event
	}

	return events, nil
}
