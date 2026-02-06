package resources

import (
	"context"
	"fmt"
	"strconv"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/config"
)

type NotificationEmail struct{}

type NotificationEmailArgs struct {
	Name         string `pulumi:"name"`
	EmailAddress string `pulumi:"emailAddress"`
}

type NotificationEmailState struct {
	NotificationEmailArgs
	LagoonID int `pulumi:"lagoonId"`
}

func (r *NotificationEmail) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "NotificationEmail")
	a.Describe(&r, "Manages a Lagoon Email notification configuration.")
}

func (a *NotificationEmailArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "The notification name.")
	an.Describe(&a.EmailAddress, "The email address to send notifications to.")
}

func (r *NotificationEmail) Create(ctx context.Context, name string, inputs NotificationEmailArgs, preview bool) (string, NotificationEmailState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if preview {
		return "preview-id", NotificationEmailState{NotificationEmailArgs: inputs}, nil
	}

	n, err := client.CreateNotificationEmail(ctx, inputs.Name, inputs.EmailAddress)
	if err != nil {
		return "", NotificationEmailState{}, fmt.Errorf("failed to create Email notification: %w", err)
	}

	return strconv.Itoa(n.ID), NotificationEmailState{NotificationEmailArgs: inputs, LagoonID: n.ID}, nil
}

func (r *NotificationEmail) Update(ctx context.Context, id string, olds NotificationEmailState, news NotificationEmailArgs, preview bool) (NotificationEmailState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	patch := map[string]any{}
	if news.EmailAddress != olds.EmailAddress {
		patch["emailAddress"] = news.EmailAddress
	}

	if preview || len(patch) == 0 {
		return NotificationEmailState{NotificationEmailArgs: news, LagoonID: olds.LagoonID}, nil
	}

	_, err := client.UpdateNotificationEmail(ctx, olds.Name, patch)
	if err != nil {
		return NotificationEmailState{}, fmt.Errorf("failed to update Email notification: %w", err)
	}

	return NotificationEmailState{NotificationEmailArgs: news, LagoonID: olds.LagoonID}, nil
}

func (r *NotificationEmail) Delete(ctx context.Context, id string, props NotificationEmailState) error {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()
	if err := client.DeleteNotificationEmail(ctx, props.Name); err != nil {
		return fmt.Errorf("failed to delete Email notification: %w", err)
	}
	return nil
}

func (r *NotificationEmail) Read(ctx context.Context, id string, inputs NotificationEmailArgs, state NotificationEmailState) (string, NotificationEmailArgs, NotificationEmailState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	name := id
	if state.Name != "" {
		name = state.Name
	}

	n, err := client.GetNotificationEmailByName(ctx, name)
	if err != nil {
		return "", NotificationEmailArgs{}, NotificationEmailState{}, fmt.Errorf("failed to read Email notification: %w", err)
	}

	args := NotificationEmailArgs{Name: n.Name, EmailAddress: n.EmailAddress}
	st := NotificationEmailState{NotificationEmailArgs: args, LagoonID: n.ID}

	return name, args, st, nil
}

func (r *NotificationEmail) Diff(ctx context.Context, id string, olds NotificationEmailState, news NotificationEmailArgs) (p.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if news.Name != olds.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if news.EmailAddress != olds.EmailAddress {
		diff["emailAddress"] = p.PropertyDiff{Kind: p.Update}
	}
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}
