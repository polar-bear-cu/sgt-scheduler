package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	KindBilling            = "billing"
	KindTrialEnd           = "trial_end"
	KindTrialEndAndBilling = "trial_end_and_billing"
)

type DueReminder struct {
	SubscriptionID  string
	UserID          string
	Name            string
	Cost            float64
	NextBillingDate time.Time
	FtEndDate       *time.Time
	BillingDue      bool
	TrialEndDue     bool
}

type ReminderJob struct {
	ReminderID     string
	UserID         string
	SubscriptionID string
	To             string
	Title          string
	Content        string
}

type SubscriptionSource interface {
	ListDueReminders(ctx context.Context, date string) ([]DueReminder, error)
}

type UserDirectory interface {
	GetEmail(ctx context.Context, userID string) (string, error)
}

type Publisher interface {
	Publish(ctx context.Context, job ReminderJob) error
}

type RunResult struct {
	Date      string
	Due       int
	Published int
	Failed    int
}

type BillingReminderUsecase struct {
	subs      SubscriptionSource
	users     UserDirectory
	pub       Publisher
	loc       *time.Location
	publicURL string
}

func NewBillingReminder(
	subs SubscriptionSource,
	users UserDirectory,
	pub Publisher,
	loc *time.Location,
	publicURL string,
) *BillingReminderUsecase {
	return &BillingReminderUsecase{subs: subs, users: users, pub: pub, loc: loc, publicURL: publicURL}
}

func (u *BillingReminderUsecase) Run(ctx context.Context, now time.Time) (RunResult, error) {
	res := RunResult{Date: now.In(u.loc).Format(time.DateOnly)}

	due, err := u.subs.ListDueReminders(ctx, res.Date)
	if err != nil {
		return res, err
	}
	res.Due = len(due)

	emails := make(map[string]string)
	var errs error
	for _, d := range due {
		email, ok := emails[d.UserID]
		if !ok {
			email, err = u.users.GetEmail(ctx, d.UserID)
			if err != nil {
				res.Failed++
				errs = errors.Join(errs, fmt.Errorf("subscription %s: get user: %w", d.SubscriptionID, err))
				continue
			}
			emails[d.UserID] = email
		}

		job := u.buildJob(d, email)
		if err := u.pub.Publish(ctx, job); err != nil {
			res.Failed++
			errs = errors.Join(errs, fmt.Errorf("reminder %s: publish: %w", job.ReminderID, err))
			continue
		}
		res.Published++
	}
	return res, errs
}

var lineBreaks = strings.NewReplacer("\r", " ", "\n", " ")

func (u *BillingReminderUsecase) buildJob(d DueReminder, email string) ReminderJob {
	name := lineBreaks.Replace(d.Name)

	kind, event := KindBilling, d.NextBillingDate
	if d.TrialEndDue && d.FtEndDate != nil {
		kind, event = KindTrialEnd, *d.FtEndDate
		if d.BillingDue {
			kind = KindTrialEndAndBilling
		}
	}
	day := event.In(u.loc).Format(time.DateOnly)
	link := u.publicURL + "/home"

	var title, content string
	switch kind {
	case KindTrialEndAndBilling:
		title = "Free trial ends and first charge: " + name
		content = fmt.Sprintf("Your free trial of %s ends on %s, and the first charge of %.2f THB is expected on the same day.\n\nReview it: %s",
			name, day, d.Cost, link)
	case KindTrialEnd:
		title = "Free trial ending: " + name
		content = fmt.Sprintf("Your free trial of %s ends on %s. After that it is expected to cost %.2f THB per billing cycle.\n\nReview it: %s",
			name, day, d.Cost, link)
	default:
		title = "Upcoming charge: " + name
		content = fmt.Sprintf("%s is expected to charge %.2f THB on %s.\n\nReview it: %s",
			name, d.Cost, day, link)
	}

	return ReminderJob{
		ReminderID:     d.SubscriptionID + ":" + kind + ":" + day,
		UserID:         d.UserID,
		SubscriptionID: d.SubscriptionID,
		To:             email,
		Title:          title,
		Content:        content,
	}
}
