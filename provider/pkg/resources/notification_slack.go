package resources

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
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

func (r *NotificationSlack) Create(ctx context.Context, req infer.CreateRequest[NotificationSlackArgs]) (infer.CreateResponse[NotificationSlackState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if req.DryRun {
		return infer.CreateResponse[NotificationSlackState]{
			ID:     "preview-id",
			Output: NotificationSlackState{NotificationSlackArgs: req.Inputs},
		}, nil
	}

	n, err := client.CreateNotificationSlack(ctx, req.Inputs.Name, req.Inputs.Webhook, req.Inputs.Channel)
	if err != nil {
		return infer.CreateResponse[NotificationSlackState]{}, fmt.Errorf("failed to create Slack notification: %w", err)
	}

	return infer.CreateResponse[NotificationSlackState]{
		ID:     strconv.Itoa(n.ID),
		Output: NotificationSlackState{NotificationSlackArgs: req.Inputs, LagoonID: n.ID},
	}, nil
}

func (r *NotificationSlack) Update(ctx context.Context, req infer.UpdateRequest[NotificationSlackArgs, NotificationSlackState]) (infer.UpdateResponse[NotificationSlackState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	patch := map[string]any{}
	if req.Inputs.Webhook != req.State.Webhook {
		patch["webhook"] = req.Inputs.Webhook
	}
	if req.Inputs.Channel != req.State.Channel {
		patch["channel"] = req.Inputs.Channel
	}

	if req.DryRun || len(patch) == 0 {
		return infer.UpdateResponse[NotificationSlackState]{
			Output: NotificationSlackState{NotificationSlackArgs: req.Inputs, LagoonID: req.State.LagoonID},
		}, nil
	}

	_, err := client.UpdateNotificationSlack(ctx, req.State.Name, patch)
	if err != nil {
		return infer.UpdateResponse[NotificationSlackState]{}, fmt.Errorf("failed to update Slack notification: %w", err)
	}

	return infer.UpdateResponse[NotificationSlackState]{
		Output: NotificationSlackState{NotificationSlackArgs: req.Inputs, LagoonID: req.State.LagoonID},
	}, nil
}

func (r *NotificationSlack) Delete(ctx context.Context, req infer.DeleteRequest[NotificationSlackState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()
	if err := c.DeleteNotificationSlack(ctx, req.State.Name); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete Slack notification: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *NotificationSlack) Read(ctx context.Context, req infer.ReadRequest[NotificationSlackArgs, NotificationSlackState]) (infer.ReadResponse[NotificationSlackArgs, NotificationSlackState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	name := req.ID
	if req.State.Name != "" {
		name = req.State.Name
	}

	n, err := c.GetNotificationSlackByName(ctx, name)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[NotificationSlackArgs, NotificationSlackState]{}, nil
		}
		return infer.ReadResponse[NotificationSlackArgs, NotificationSlackState]{}, fmt.Errorf("failed to read Slack notification: %w", err)
	}

	args := NotificationSlackArgs{Name: n.Name, Webhook: n.Webhook, Channel: n.Channel}
	st := NotificationSlackState{NotificationSlackArgs: args, LagoonID: n.ID}

	return infer.ReadResponse[NotificationSlackArgs, NotificationSlackState]{
		ID:     strconv.Itoa(n.ID),
		Inputs: args,
		State:  st,
	}, nil
}

func (r *NotificationSlack) Diff(ctx context.Context, req infer.DiffRequest[NotificationSlackArgs, NotificationSlackState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.Inputs.Name != req.State.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.Inputs.Webhook != req.State.Webhook {
		diff["webhook"] = p.PropertyDiff{Kind: p.Update}
	}
	if req.Inputs.Channel != req.State.Channel {
		diff["channel"] = p.PropertyDiff{Kind: p.Update}
	}
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}
