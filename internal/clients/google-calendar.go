package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	maxResults = 100
)

type GoogleCalendar struct {
	oauthConfig *oauth2.Config
}

type GoogleCalendarCfg struct {
	ClientId     string `env:"GOOGLE_CALENDAR_CLIENT_ID" env-required:"true" yaml:"client-id"`
	ClientSecret string `env:"GOOGLE_CALENDAR_CLIENT_SECRET" env-required:"true" yaml:"client-secret"`
	RedirectURL  string `env:"GOOGLE_CALENDAR_REDIRECT_URL" env-required:"true" yaml:"redirect-url"`
}

func New(ctx context.Context, cfg GoogleCalendarCfg) *GoogleCalendar {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientId,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}

	return &GoogleCalendar{
		oauthCfg,
	}
}

// return redirect URL
func (g *GoogleCalendar) LoginURL(ctx context.Context, state string) string {
	return g.oauthConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)
}

func (g *GoogleCalendar) GetTokenFromCode(ctx context.Context, authCode string) (models.Token, error) {
	const op = "google-calendar.GetTokenFromCode"

	tok, err := g.oauthConfig.Exchange(ctx, authCode)
	if err != nil {
		return models.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.Token{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
	}, nil
}

func (g *GoogleCalendar) GetEvents(ctx context.Context, tok models.Token, minTime, maxTime time.Time) (*[]*models.CalendarEvent, error) {
	const op = "google-calebdar.GetEvents"

	srv, err := g.serviceFromToken(ctx, tok)
	if err != nil {
		err = HandleGoogleAPIError(err)

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	calendars, err := srv.CalendarList.List().Do()
	if err != nil {
		err = HandleGoogleAPIError(err)

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	events := make([]*models.CalendarEvent, 0)
	for _, item := range calendars.Items {
		eventsOfCalendar, err := srv.Events.List(item.Id).
			ShowDeleted(false).
			SingleEvents(true).
			TimeMin(minTime.Format(time.RFC3339)).
			TimeMax(maxTime.Format(time.RFC3339)).
			MaxResults(maxResults).OrderBy("startTime").
			Do()
		if err != nil {
			err = HandleGoogleAPIError(err)

			return nil, fmt.Errorf("%s: %w", op, err)
		}

		for _, event := range eventsOfCalendar.Items {
			st, err := time.Parse(time.RFC3339, event.Start.DateTime)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			end, err := time.Parse(time.RFC3339, event.End.DateTime)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			events = append(events, &models.CalendarEvent{
				Title:      event.Summary,
				EventId:    event.Id,
				CalendarId: item.Id,
				Start:      st,
				End:        end,
			})
		}
	}

	return &events, nil
}

func (g *GoogleCalendar) CreateEvent(ctx context.Context, tok models.Token, start, end time.Time, eventId, calendarId string) error {
	const op = "google-calendar.CreateEvent"

	srv, err := g.serviceFromToken(ctx, tok)
	if err != nil {
		err = HandleGoogleAPIError(err)

		return fmt.Errorf("%s: failed to get token: %w", op, err)
	}

	event := calendar.Event{
		Summary: eventId,
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
		},
	}

	_, err = srv.Events.Insert(calendarId, &event).Do()
	if err != nil {
		err = HandleGoogleAPIError(err)

		return fmt.Errorf("%s: failed to create event: %w", op, err)
	}

	return nil
}

func (g *GoogleCalendar) DeleteEvent(ctx context.Context, tok models.Token, eventId, calendarId string) error {
	const op = "google-calendar.DeleteEvent"

	srv, err := g.serviceFromToken(ctx, tok)
	if err != nil {
		err = HandleGoogleAPIError(err)

		return fmt.Errorf("%s: %w", op, err)
	}

	err = srv.Events.Delete(calendarId, eventId).Do()
	if err != nil {
		err = HandleGoogleAPIError(err)

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (g *GoogleCalendar) CreateCalendar(ctx context.Context, tok models.Token, title string) (*models.Calendar, error) {
	const op = "google-calendar.CreateCalendar"

	srv, err := g.serviceFromToken(ctx, tok)
	if err != nil {
		err = HandleGoogleAPIError(err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	calend := calendar.Calendar{
		Summary: title,
	}
	craetedCal, err := srv.Calendars.Insert(&calend).Do()
	if err != nil {
		err = HandleGoogleAPIError(err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.Calendar{
		Title: title,
		Id:    craetedCal.Id,
	}, nil
}

func (g *GoogleCalendar) serviceFromToken(ctx context.Context, tok models.Token) (*calendar.Service, error) {
	const op = "google-calendar.ServiceFromToken"

	authToken, err := g.TokenConvert(ctx, tok)
	if err != nil {
		err = HandleGoogleAPIError(err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	client := g.oauthConfig.Client(ctx, authToken)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		err = HandleGoogleAPIError(err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return srv, nil
}

func (g *GoogleCalendar) RefreshToken(ctx context.Context, tok models.Token) (models.Token, error) {
	const op = "google-calendar.RefreshToken"

	newToken, err := g.oauthConfig.TokenSource(ctx, &oauth2.Token{RefreshToken: tok.RefreshToken}).Token()
	if err != nil {
		err = HandleGoogleAPIError(err)
		return models.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.Token{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
	}, nil
}

func (g *GoogleCalendar) TokenConvert(ctx context.Context, tok models.Token) (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
	}

	return g.oauthConfig.TokenSource(ctx, token).Token()
}
