package resources

import (
	"context"
	"fmt"
	"strconv"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
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

func (r *NotificationMicrosoftTeams) Create(ctx context.Context, name string, inputs NotificationMicrosoftTeamsArgs, preview bool) (string, NotificationMicrosoftTeamsState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if preview {
		return "preview-id", NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: inputs}, nil
	}

	n, err := client.CreateNotificationMicrosoftTeams(ctx, inputs.Name, inputs.Webhook)
	if err != nil {
		return "", NotificationMicrosoftTeamsState{}, fmt.Errorf("failed to create Microsoft Teams notification: %w", err)
	}

	return strconv.Itoa(n.ID), NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: inputs, LagoonID: n.ID}, nil
}

func (r *NotificationMicrosoftTeams) Update(ctx context.Context, id string, olds NotificationMicrosoftTeamsState, news NotificationMicrosoftTeamsArgs, preview bool) (NotificationMicrosoftTeamsState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	patch := map[string]any{}
	if news.Webhook != olds.Webhook {
		patch["webhook"] = news.Webhook
	}

	if preview || len(patch) == 0 {
		return NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: news, LagoonID: olds.LagoonID}, nil
	}

	_, err := client.UpdateNotificationMicrosoftTeams(ctx, olds.Name, patch)
	if err != nil {
		return NotificationMicrosoftTeamsState{}, fmt.Errorf("failed to update Microsoft Teams notification: %w", err)
	}

	return NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: news, LagoonID: olds.LagoonID}, nil
}

func (r *NotificationMicrosoftTeams) Delete(ctx context.Context, id string, props NotificationMicrosoftTeamsState) error {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()
	if err := client.DeleteNotificationMicrosoftTeams(ctx, props.Name); err != nil {
		return fmt.Errorf("failed to delete Microsoft Teams notification: %w", err)
	}
	return nil
}

func (r *NotificationMicrosoftTeams) Read(ctx context.Context, id string, inputs NotificationMicrosoftTeamsArgs, state NotificationMicrosoftTeamsState) (string, NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	name := id
	if state.Name != "" {
		name = state.Name
	}

	n, err := client.GetNotificationMicrosoftTeamsByName(ctx, name)
	if err != nil {
		return "", NotificationMicrosoftTeamsArgs{}, NotificationMicrosoftTeamsState{}, fmt.Errorf("failed to read Microsoft Teams notification: %w", err)
	}

	args := NotificationMicrosoftTeamsArgs{Name: n.Name, Webhook: n.Webhook}
	st := NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: args, LagoonID: n.ID}

	return name, args, st, nil
}

func (r *NotificationMicrosoftTeams) Diff(ctx context.Context, id string, olds NotificationMicrosoftTeamsState, news NotificationMicrosoftTeamsArgs) (p.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if news.Name != olds.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if news.Webhook != olds.Webhook {
		diff["webhook"] = p.PropertyDiff{Kind: p.Update}
	}
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}
