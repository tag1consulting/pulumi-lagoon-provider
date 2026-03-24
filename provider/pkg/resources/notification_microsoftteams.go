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

type NotificationMicrosoftTeams struct{}

type NotificationMicrosoftTeamsArgs struct {
	Name    string `pulumi:"name"`
	Webhook string `pulumi:"webhook" provider:"secret"`
}

type NotificationMicrosoftTeamsState struct {
	NotificationMicrosoftTeamsArgs
	LagoonID int `pulumi:"lagoonId"`
}

func (r *NotificationMicrosoftTeams) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "NotificationMicrosoftTeams")
	a.Describe(&r, "Manages a Lagoon Microsoft Teams notification configuration.")
}

func (a *NotificationMicrosoftTeamsArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "The notification name.")
	an.Describe(&a.Webhook, "The Microsoft Teams webhook URL (stored as secret).")
}

func (r *NotificationMicrosoftTeams) Create(ctx context.Context, req infer.CreateRequest[NotificationMicrosoftTeamsArgs]) (infer.CreateResponse[NotificationMicrosoftTeamsState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if req.DryRun {
		return infer.CreateResponse[NotificationMicrosoftTeamsState]{
			ID:     "preview-id",
			Output: NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: req.Inputs},
		}, nil
	}

	n, err := client.CreateNotificationMicrosoftTeams(ctx, req.Inputs.Name, req.Inputs.Webhook)
	if err != nil {
		return infer.CreateResponse[NotificationMicrosoftTeamsState]{}, fmt.Errorf("failed to create Microsoft Teams notification: %w", err)
	}

	return infer.CreateResponse[NotificationMicrosoftTeamsState]{
		ID:     strconv.Itoa(n.ID),
		Output: NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: req.Inputs, LagoonID: n.ID},
	}, nil
}

func (r *NotificationMicrosoftTeams) Update(ctx context.Context, req infer.UpdateRequest[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]) (infer.UpdateResponse[NotificationMicrosoftTeamsState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	patch := map[string]any{}
	if req.Inputs.Webhook != req.State.Webhook {
		patch["webhook"] = req.Inputs.Webhook
	}

	if req.DryRun || len(patch) == 0 {
		return infer.UpdateResponse[NotificationMicrosoftTeamsState]{
			Output: NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: req.Inputs, LagoonID: req.State.LagoonID},
		}, nil
	}

	_, err := client.UpdateNotificationMicrosoftTeams(ctx, req.State.Name, patch)
	if err != nil {
		return infer.UpdateResponse[NotificationMicrosoftTeamsState]{}, fmt.Errorf("failed to update Microsoft Teams notification: %w", err)
	}

	return infer.UpdateResponse[NotificationMicrosoftTeamsState]{
		Output: NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: req.Inputs, LagoonID: req.State.LagoonID},
	}, nil
}

func (r *NotificationMicrosoftTeams) Delete(ctx context.Context, req infer.DeleteRequest[NotificationMicrosoftTeamsState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()
	if err := c.DeleteNotificationMicrosoftTeams(ctx, req.State.Name); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete Microsoft Teams notification: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *NotificationMicrosoftTeams) Read(ctx context.Context, req infer.ReadRequest[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]) (infer.ReadResponse[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	// Prefer state name for lookup, fallback to ID
	name := req.State.Name
	if name == "" {
		name = req.ID
	}

	n, err := c.GetNotificationMicrosoftTeamsByName(ctx, name)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]{}, nil
		}
		return infer.ReadResponse[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]{}, fmt.Errorf("failed to read Microsoft Teams notification: %w", err)
	}

	args := NotificationMicrosoftTeamsArgs{Name: n.Name, Webhook: n.Webhook}
	st := NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: args, LagoonID: n.ID}

	return infer.ReadResponse[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]{
		ID:     strconv.Itoa(n.ID),
		Inputs: args,
		State:  st,
	}, nil
}

func (r *NotificationMicrosoftTeams) Diff(ctx context.Context, req infer.DiffRequest[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.Inputs.Name != req.State.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.Inputs.Webhook != req.State.Webhook {
		diff["webhook"] = p.PropertyDiff{Kind: p.Update}
	}
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}
