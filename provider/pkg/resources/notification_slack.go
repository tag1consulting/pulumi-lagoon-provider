package resources

import (
	"context"
	"fmt"
	"strconv"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/config"
)

// NotificationSlack manages a Lagoon Slack notification.
type NotificationSlack struct{}

type NotificationSlackArgs struct {
	Name    string `pulumi:"name"`
	Webhook string `pulumi:"webhook" provider:"secret"`
	Channel string `pulumi:"channel"`
}

type NotificationSlackState struct {
	NotificationSlackArgs
	LagoonID int `pulumi:"lagoonId"`
}

func (r *NotificationSlack) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "NotificationSlack")
	a.Describe(&r, "Manages a Lagoon Slack notification configuration.")
}

func (a *NotificationSlackArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "The notification name.")
	an.Describe(&a.Webhook, "The Slack webhook URL (stored as secret).")
	an.Describe(&a.Channel, "The Slack channel (e.g., '#deployments').")
}

func (r *NotificationSlack) Create(ctx context.Context, name string, inputs NotificationSlackArgs, preview bool) (string, NotificationSlackState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if preview {
		return "preview-id", NotificationSlackState{NotificationSlackArgs: inputs}, nil
	}

	n, err := client.CreateNotificationSlack(ctx, inputs.Name, inputs.Webhook, inputs.Channel)
	if err != nil {
		return "", NotificationSlackState{}, fmt.Errorf("failed to create Slack notification: %w", err)
	}

	return strconv.Itoa(n.ID), NotificationSlackState{NotificationSlackArgs: inputs, LagoonID: n.ID}, nil
}

func (r *NotificationSlack) Update(ctx context.Context, id string, olds NotificationSlackState, news NotificationSlackArgs, preview bool) (NotificationSlackState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	patch := map[string]any{}
	if news.Webhook != olds.Webhook {
		patch["webhook"] = news.Webhook
	}
	if news.Channel != olds.Channel {
		patch["channel"] = news.Channel
	}

	if preview || len(patch) == 0 {
		return NotificationSlackState{NotificationSlackArgs: news, LagoonID: olds.LagoonID}, nil
	}

	_, err := client.UpdateNotificationSlack(ctx, olds.Name, patch)
	if err != nil {
		return NotificationSlackState{}, fmt.Errorf("failed to update Slack notification: %w", err)
	}

	return NotificationSlackState{NotificationSlackArgs: news, LagoonID: olds.LagoonID}, nil
}

func (r *NotificationSlack) Delete(ctx context.Context, id string, props NotificationSlackState) error {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()
	if err := client.DeleteNotificationSlack(ctx, props.Name); err != nil {
		return fmt.Errorf("failed to delete Slack notification: %w", err)
	}
	return nil
}

func (r *NotificationSlack) Read(ctx context.Context, id string, inputs NotificationSlackArgs, state NotificationSlackState) (string, NotificationSlackArgs, NotificationSlackState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	name := id
	if state.Name != "" {
		name = state.Name
	}

	n, err := client.GetNotificationSlackByName(ctx, name)
	if err != nil {
		return "", NotificationSlackArgs{}, NotificationSlackState{}, fmt.Errorf("failed to read Slack notification: %w", err)
	}

	args := NotificationSlackArgs{Name: n.Name, Webhook: n.Webhook, Channel: n.Channel}
	st := NotificationSlackState{NotificationSlackArgs: args, LagoonID: n.ID}

	return name, args, st, nil
}

func (r *NotificationSlack) Diff(ctx context.Context, id string, olds NotificationSlackState, news NotificationSlackArgs) (p.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if news.Name != olds.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if news.Webhook != olds.Webhook {
		diff["webhook"] = p.PropertyDiff{Kind: p.Update}
	}
	if news.Channel != olds.Channel {
		diff["channel"] = p.PropertyDiff{Kind: p.Update}
	}
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}
