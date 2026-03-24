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

func (r *NotificationEmail) Create(ctx context.Context, req infer.CreateRequest[NotificationEmailArgs]) (infer.CreateResponse[NotificationEmailState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if req.DryRun {
		return infer.CreateResponse[NotificationEmailState]{
			ID:     "preview-id",
			Output: NotificationEmailState{NotificationEmailArgs: req.Inputs},
		}, nil
	}

	n, err := client.CreateNotificationEmail(ctx, req.Inputs.Name, req.Inputs.EmailAddress)
	if err != nil {
		return infer.CreateResponse[NotificationEmailState]{}, fmt.Errorf("failed to create Email notification: %w", err)
	}

	return infer.CreateResponse[NotificationEmailState]{
		ID:     strconv.Itoa(n.ID),
		Output: NotificationEmailState{NotificationEmailArgs: req.Inputs, LagoonID: n.ID},
	}, nil
}

func (r *NotificationEmail) Update(ctx context.Context, req infer.UpdateRequest[NotificationEmailArgs, NotificationEmailState]) (infer.UpdateResponse[NotificationEmailState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	patch := map[string]any{}
	if req.Inputs.EmailAddress != req.State.EmailAddress {
		patch["emailAddress"] = req.Inputs.EmailAddress
	}

	if req.DryRun || len(patch) == 0 {
		return infer.UpdateResponse[NotificationEmailState]{
			Output: NotificationEmailState{NotificationEmailArgs: req.Inputs, LagoonID: req.State.LagoonID},
		}, nil
	}

	_, err := client.UpdateNotificationEmail(ctx, req.State.Name, patch)
	if err != nil {
		return infer.UpdateResponse[NotificationEmailState]{}, fmt.Errorf("failed to update Email notification: %w", err)
	}

	return infer.UpdateResponse[NotificationEmailState]{
		Output: NotificationEmailState{NotificationEmailArgs: req.Inputs, LagoonID: req.State.LagoonID},
	}, nil
}

func (r *NotificationEmail) Delete(ctx context.Context, req infer.DeleteRequest[NotificationEmailState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()
	if err := c.DeleteNotificationEmail(ctx, req.State.Name); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete Email notification: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *NotificationEmail) Read(ctx context.Context, req infer.ReadRequest[NotificationEmailArgs, NotificationEmailState]) (infer.ReadResponse[NotificationEmailArgs, NotificationEmailState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	// Prefer state name for lookup, fallback to ID
	name := req.State.Name
	if name == "" {
		name = req.ID
	}

	n, err := c.GetNotificationEmailByName(ctx, name)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[NotificationEmailArgs, NotificationEmailState]{}, nil
		}
		return infer.ReadResponse[NotificationEmailArgs, NotificationEmailState]{}, fmt.Errorf("failed to read Email notification: %w", err)
	}

	args := NotificationEmailArgs{Name: n.Name, EmailAddress: n.EmailAddress}
	st := NotificationEmailState{NotificationEmailArgs: args, LagoonID: n.ID}

	return infer.ReadResponse[NotificationEmailArgs, NotificationEmailState]{
		ID:     strconv.Itoa(n.ID),
		Inputs: args,
		State:  st,
	}, nil
}

func (r *NotificationEmail) Diff(ctx context.Context, req infer.DiffRequest[NotificationEmailArgs, NotificationEmailState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.Inputs.Name != req.State.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.Inputs.EmailAddress != req.State.EmailAddress {
		diff["emailAddress"] = p.PropertyDiff{Kind: p.Update}
	}
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}
