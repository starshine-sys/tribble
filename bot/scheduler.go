// SPDX-License-Identifier: AGPL-3.0-only
package bot

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/starshine-sys/tribble/db"
)

// Event is any event that can be scheduled.
// Execute is called when the event is due to fire, Offset is called to determine how much to move the event.
type Event interface {
	Execute(ctx context.Context, id int64, bot *Bot) error
	Offset() time.Duration
}

// Return Reschedule if the event should be rescheduled (offset by the duration returned from Offset)
const Reschedule = errors.Sentinel("reschedule event")

type Scheduler struct {
	*db.DB

	started bool

	bot    *Bot
	mu     sync.RWMutex
	events map[string]func() Event
}

func NewScheduler(bot *Bot, db *db.DB) *Scheduler {
	return &Scheduler{
		DB:     db,
		bot:    bot,
		events: map[string]func() Event{},
	}
}

// AddType adds event common.
// These should be *pointers*, anything else will panic, even if it implements Event!
func (s *Scheduler) AddType(events ...Event) {
	s.mu.Lock()
	for _, v := range events {
		t := reflect.ValueOf(v).Elem().Type()

		s.bot.Sugar.Infof("Adding type %q to scheduler", t.String())

		s.events[t.String()] = func() Event {
			return reflect.New(t).Interface().(Event)
		}
	}
	s.mu.Unlock()
}

func (s *Scheduler) Add(t time.Time, v Event) (id int64, err error) {
	typ := reflect.ValueOf(v).Elem().Type()

	s.mu.RLock()
	_, ok := s.events[typ.String()]
	s.mu.RUnlock()
	if !ok {
		return 0, ErrUnknownEvent
	}

	s.bot.Sugar.Debugf("Scheduling event type %q for %v", typ.String(), t)

	b, err := json.Marshal(v)
	if err != nil {
		return 0, errors.Wrap(err, "marshal json")
	}

	return id, s.Pool.QueryRow(context.Background(), "insert into public.scheduled_events (event_type, expires, data) values ($1, $2, $3) returning id", typ.String(), t.UTC(), b).Scan(&id)
}

func (s *Scheduler) Remove(id int64) error {
	_, err := s.Pool.Exec(context.Background(), "delete from public.scheduled_events where id = $1", id)
	return err
}

func (s *Scheduler) Reschedule(id int64, dur time.Duration) error {
	_, err := s.Pool.Exec(context.Background(), "update public.scheduled_events set expires = $2 where id = $1", id, time.Now().UTC().Add(dur))
	return err
}

type row struct {
	ID        int64
	EventType string
	Expires   time.Time
	Data      json.RawMessage
}

func (s *Scheduler) expiring(ctx context.Context) ([]row, error) {
	var rs []row

	err := pgxscan.Select(ctx, s.Pool, &rs, "select * from public.scheduled_events where expires < current_timestamp at time zone 'utc' order by id asc limit 5")
	return rs, errors.Cause(err)
}

// Start starts the scheduler. *This function is blocking!*
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.started {
		s.bot.Sugar.Warnf("Scheduler.Start called after the scheduler was already started")
		s.mu.Unlock()
		return
	}
	s.started = true
	s.mu.Unlock()

	s.bot.Sugar.Infof("Starting scheduler")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			s.bot.Sugar.Info("Stopping scheduler")
			return

		case <-ticker.C:
		}

		err := s.tick(ctx)
		if err != nil {
			s.bot.Sugar.Errorf("Error running scheduler tick: %v", err)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) error {
	rs, err := s.expiring(ctx)
	if err != nil {
		return err
	}

	for _, r := range rs {
		dur, err := s.run(r)
		if err != nil {
			if err == Reschedule {
				err = s.Reschedule(r.ID, dur)
				if err != nil {
					return errors.Wrap(err, "rescheduling event")
				}
			} else {
				err = s.Remove(r.ID)
				if err != nil {
					return errors.Wrap(err, "removing errored event")
				}
			}
		} else {
			// otherwise, remove the event, as it's done
			err = s.Remove(r.ID)
			if err != nil {
				return errors.Wrap(err, "removing completed event")
			}
		}
	}

	return err
}

const ErrUnknownEvent = errors.Sentinel("unknown event type")

func (s *Scheduler) run(r row) (offset time.Duration, err error) {
	var ev Event
	s.mu.RLock()
	fn, ok := s.events[r.EventType]
	s.mu.RUnlock()
	if !ok {
		return 0, ErrUnknownEvent
	}
	ev = fn()

	err = json.Unmarshal(r.Data, ev)
	if err != nil {
		return 0, errors.Wrap(err, "unmarshaling json")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = ev.Execute(ctx, r.ID, s.bot)
	if err != nil {
		return ev.Offset(), err
	}
	return 0, nil
}
