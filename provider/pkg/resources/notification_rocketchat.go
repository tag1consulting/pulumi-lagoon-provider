package resources

import (
	"context"
	"fmt"
	"strconv"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/config"
)

type NotificationRocketChat struct{}

type NotificationRocketChatArgs struct {
	Name    string `pulumi:"name"`
	Webhook string `pulumi:"webhook" provider:"secret"`
	Channel string `pulumi:"channel"`
}

type NotificationRocketChatState struct {
	NotificationRocketChatArgs
	LagoonID int `pulumi:"lagoonId"`
}

func (r *NotificationRocketChat) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "NotificationRocketChat")
	a.Describe(&r, "Manages a Lagoon RocketChat notification configuration.")
}

func (a *NotificationRocketChatArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "The notification name.")
	an.Describe(&a.Webhook, "The RocketChat webhook URL (stored as secret).")
	an.Describe(&a.Channel, "The RocketChat channel.")
}

func (r *NotificationRocketChat) Create(ctx context.Context, name string, inputs NotificationRocketChatArgs, preview bool) (string, NotificationRocketChatState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if preview {
		return "preview-id", NotificationRocketChatState{NotificationRocketChatArgs: inputs}, nil
	}

	n, err := client.CreateNotificationRocketChat(ctx, inputs.Name, inputs.Webhook, inputs.Channel)
	if err != nil {
		return "", NotificationRocketChatState{}, fmt.Errorf("failed to create RocketChat notification: %w", err)
	}

	return strconv.Itoa(n.ID), NotificationRocketChatState{NotificationRocketChatArgs: inputs, LagoonID: n.ID}, nil
}

func (r *NotificationRocketChat) Update(ctx context.Context, id string, olds NotificationRocketChatState, news NotificationRocketChatArgs, preview bool) (NotificationRocketChatState, error) {
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
		return NotificationRocketChatState{NotificationRocketChatArgs: news, LagoonID: olds.LagoonID}, nil
	}

	_, err := client.UpdateNotificationRocketChat(ctx, olds.Name, patch)
	if err != nil {
		return NotificationRocketChatState{}, fmt.Errorf("failed to update RocketChat notification: %w", err)
	}

	return NotificationRocketChatState{NotificationRocketChatArgs: news, LagoonID: olds.LagoonID}, nil
}

func (r *NotificationRocketChat) Delete(ctx context.Context, id string, props NotificationRocketChatState) error {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()
	if err := client.DeleteNotificationRocketChat(ctx, props.Name); err != nil {
		return fmt.Errorf("failed to delete RocketChat notification: %w", err)
	}
	return nil
}

func (r *NotificationRocketChat) Read(ctx context.Context, id string, inputs NotificationRocketChatArgs, state NotificationRocketChatState) (string, NotificationRocketChatArgs, NotificationRocketChatState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	name := id
	if state.Name != "" {
		name = state.Name
	}

	n, err := client.GetNotificationRocketChatByName(ctx, name)
	if err != nil {
		return "", NotificationRocketChatArgs{}, NotificationRocketChatState{}, fmt.Errorf("failed to read RocketChat notification: %w", err)
	}

	args := NotificationRocketChatArgs{Name: n.Name, Webhook: n.Webhook, Channel: n.Channel}
	st := NotificationRocketChatState{NotificationRocketChatArgs: args, LagoonID: n.ID}

	return name, args, st, nil
}

func (r *NotificationRocketChat) Diff(ctx context.Context, id string, olds NotificationRocketChatState, news NotificationRocketChatArgs) (p.DiffResponse, error) {
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
