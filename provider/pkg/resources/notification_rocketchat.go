package resources

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
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

func (r *NotificationRocketChat) Create(ctx context.Context, req infer.CreateRequest[NotificationRocketChatArgs]) (infer.CreateResponse[NotificationRocketChatState], error) {
	client := clientFor(ctx)

	if req.DryRun {
		return infer.CreateResponse[NotificationRocketChatState]{
			ID:     "preview-id",
			Output: NotificationRocketChatState{NotificationRocketChatArgs: req.Inputs},
		}, nil
	}

	n, err := client.CreateNotificationRocketChat(ctx, req.Inputs.Name, req.Inputs.Webhook, req.Inputs.Channel)
	if err != nil {
		return infer.CreateResponse[NotificationRocketChatState]{}, fmt.Errorf("failed to create RocketChat notification: %w", err)
	}

	return infer.CreateResponse[NotificationRocketChatState]{
		ID:     strconv.Itoa(n.ID),
		Output: NotificationRocketChatState{NotificationRocketChatArgs: req.Inputs, LagoonID: n.ID},
	}, nil
}

func (r *NotificationRocketChat) Update(ctx context.Context, req infer.UpdateRequest[NotificationRocketChatArgs, NotificationRocketChatState]) (infer.UpdateResponse[NotificationRocketChatState], error) {
	client := clientFor(ctx)

	patch := map[string]any{}
	if req.Inputs.Webhook != req.State.Webhook {
		patch["webhook"] = req.Inputs.Webhook
	}
	if req.Inputs.Channel != req.State.Channel {
		patch["channel"] = req.Inputs.Channel
	}

	if req.DryRun || len(patch) == 0 {
		return infer.UpdateResponse[NotificationRocketChatState]{
			Output: NotificationRocketChatState{NotificationRocketChatArgs: req.Inputs, LagoonID: req.State.LagoonID},
		}, nil
	}

	_, err := client.UpdateNotificationRocketChat(ctx, req.State.Name, patch)
	if err != nil {
		return infer.UpdateResponse[NotificationRocketChatState]{}, fmt.Errorf("failed to update RocketChat notification: %w", err)
	}

	return infer.UpdateResponse[NotificationRocketChatState]{
		Output: NotificationRocketChatState{NotificationRocketChatArgs: req.Inputs, LagoonID: req.State.LagoonID},
	}, nil
}

func (r *NotificationRocketChat) Delete(ctx context.Context, req infer.DeleteRequest[NotificationRocketChatState]) (infer.DeleteResponse, error) {
	c := clientFor(ctx)
	if err := c.DeleteNotificationRocketChat(ctx, req.State.Name); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete RocketChat notification: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *NotificationRocketChat) Read(ctx context.Context, req infer.ReadRequest[NotificationRocketChatArgs, NotificationRocketChatState]) (infer.ReadResponse[NotificationRocketChatArgs, NotificationRocketChatState], error) {
	c := clientFor(ctx)

	// Prefer state name for lookup, fallback to ID
	name := req.State.Name
	if name == "" {
		name = req.ID
	}

	n, err := c.GetNotificationRocketChatByName(ctx, name)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[NotificationRocketChatArgs, NotificationRocketChatState]{}, nil
		}
		return infer.ReadResponse[NotificationRocketChatArgs, NotificationRocketChatState]{}, fmt.Errorf("failed to read RocketChat notification: %w", err)
	}

	args := NotificationRocketChatArgs{Name: n.Name, Webhook: n.Webhook, Channel: n.Channel}
	st := NotificationRocketChatState{NotificationRocketChatArgs: args, LagoonID: n.ID}

	return infer.ReadResponse[NotificationRocketChatArgs, NotificationRocketChatState]{
		ID:     strconv.Itoa(n.ID),
		Inputs: args,
		State:  st,
	}, nil
}

func (r *NotificationRocketChat) Diff(ctx context.Context, req infer.DiffRequest[NotificationRocketChatArgs, NotificationRocketChatState]) (infer.DiffResponse, error) {
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