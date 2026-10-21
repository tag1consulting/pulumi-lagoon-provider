package provider

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/blang/semver"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/integration"
	presource "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/resources"
)

// These tests drive the real infer layer in preview (dry run) mode. A unit test of the
// resource methods cannot see how infer marks outputs unknown, which is what caused the
// replace cascade in issue #305.

func newPreviewServer(t *testing.T) integration.Server {
	t.Helper()
	prov, err := NewProvider("0.0.0-test")
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	srv, err := integration.NewServer(context.Background(), "lagoon", semver.MustParse("0.0.1"), integration.WithProvider(prov))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	// Dry runs never call the Lagoon API, so the endpoint is never contacted.
	err = srv.Configure(p.ConfigureRequest{Args: property.NewMap(map[string]property.Value{
		"apiUrl": property.New("http://127.0.0.1:1/graphql"),
		"token":  property.New("test-token"),
	})})
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}
	return srv
}

func previewURN(typ string) presource.URN {
	return presource.CreateURN("res", "lagoon:lagoon:"+typ, "", "proj", "stack")
}

func pmap(kv map[string]any) property.Map {
	m := map[string]property.Value{}
	for k, v := range kv {
		switch x := v.(type) {
		case property.Value:
			m[k] = x
		case string:
			m[k] = property.New(x)
		case int:
			m[k] = property.New(float64(x))
		default:
			panic("unsupported test value")
		}
	}
	return property.NewMap(m)
}

// merged returns a copy of base with the overrides applied.
func merged(base map[string]any, overrides map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range overrides {
		out[k] = v
	}
	return out
}

var unknown = property.New(property.Computed)

type previewCase struct {
	name     string
	typ      string
	inputs   map[string]any // old inputs
	outputs  map[string]any // outputs in the old state, in addition to the inputs
	change   map[string]any // changes applied to the inputs for the preview
	known    []string       // outputs that must stay known
	computed []string       // outputs that must be unknown
}

func previewCases() []previewCase {
	dt := map[string]any{"name": "dt", "consoleUrl": "https://k8s.example.com", "buildImage": "img:1"}
	dtOut := map[string]any{"lagoonId": 1, "created": "2026-01-01"}
	proj := map[string]any{"name": "proj", "gitUrl": "git@example.com:r.git", "deploytargetId": 1}
	projOut := map[string]any{"lagoonId": 2, "publicKey": "ssh-rsa AAAA", "created": "2026-01-01"}
	// markComputed only sees a changed input when it had an old value, so deployTitle is set here.
	env := map[string]any{"name": "main", "projectId": 2, "deployType": "branch", "environmentType": "production", "deployTitle": "old"}
	envOut := map[string]any{"lagoonId": 3, "route": "https://a", "routes": "https://a", "created": "2026-01-01"}

	return []previewCase{
		{name: "DeployTarget unrelated change keeps lagoonId", typ: "DeployTarget", inputs: dt, outputs: dtOut,
			change: map[string]any{"buildImage": "img:2"}, known: []string{"lagoonId"}, computed: []string{"created"}},
		{name: "DeployTarget rename makes lagoonId unknown", typ: "DeployTarget", inputs: dt, outputs: dtOut,
			change: map[string]any{"name": "dt2"}, computed: []string{"lagoonId", "created"}},

		{name: "Project unrelated change keeps lagoonId", typ: "Project", inputs: proj, outputs: projOut,
			change: map[string]any{"gitUrl": "git@example.com:other.git"}, known: []string{"lagoonId"}, computed: []string{"publicKey", "created"}},
		{name: "Project with unknown deploy target keeps lagoonId (issue 305 cascade)", typ: "Project", inputs: proj, outputs: projOut,
			change: map[string]any{"deploytargetId": unknown}, known: []string{"lagoonId"}, computed: []string{"publicKey", "created"}},
		{name: "Project rename makes lagoonId unknown", typ: "Project", inputs: proj, outputs: projOut,
			change: map[string]any{"name": "proj2"}, computed: []string{"lagoonId", "publicKey", "created"}},

		{name: "Environment unrelated change keeps lagoonId", typ: "Environment", inputs: env, outputs: envOut,
			change: map[string]any{"deployTitle": "new"}, known: []string{"lagoonId"}, computed: []string{"route", "routes", "created"}},
		{name: "Environment rename makes lagoonId unknown", typ: "Environment", inputs: env, outputs: envOut,
			change: map[string]any{"name": "dev"}, computed: []string{"lagoonId", "route", "routes", "created"}},
		{name: "Environment with unknown project makes lagoonId unknown", typ: "Environment", inputs: env, outputs: envOut,
			change: map[string]any{"projectId": unknown}, computed: []string{"lagoonId", "route", "routes", "created"}},
	}
}

func TestPreviewUpdate_LagoonIDStaysKnownOnUnrelatedChange(t *testing.T) {
	srv := newPreviewServer(t)
	for _, tc := range previewCases() {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := srv.Update(p.UpdateRequest{
				ID:        "1",
				Urn:       previewURN(tc.typ),
				State:     pmap(merged(tc.inputs, tc.outputs)),
				OldInputs: pmap(tc.inputs),
				Inputs:    pmap(merged(tc.inputs, tc.change)),
				DryRun:    true,
			})
			if err != nil {
				t.Fatalf("Update: %v", err)
			}
			for _, k := range tc.known {
				if v := resp.Properties.Get(k); v.IsComputed() {
					t.Errorf("%s must stay known in the preview, got unknown", k)
				}
			}
			for _, k := range tc.computed {
				if v := resp.Properties.Get(k); !v.IsComputed() {
					t.Errorf("%s must be unknown in the preview, got %v", k, v)
				}
			}
		})
	}
}

func TestPreviewCreate_LagoonIDIsUnknown(t *testing.T) {
	srv := newPreviewServer(t)
	creates := map[string]map[string]any{
		"DeployTarget": {"name": "dt", "consoleUrl": "https://k8s.example.com"},
		"Project":      {"name": "proj", "gitUrl": "git@example.com:r.git", "deploytargetId": unknown},
		"Environment":  {"name": "main", "projectId": unknown, "deployType": "branch", "environmentType": "production"},
	}
	for typ, inputs := range creates {
		t.Run(typ, func(t *testing.T) {
			resp, err := srv.Create(p.CreateRequest{Urn: previewURN(typ), Properties: pmap(inputs), DryRun: true})
			if err != nil {
				t.Fatalf("Create: %v", err)
			}
			if v := resp.Properties.Get("lagoonId"); !v.IsComputed() {
				t.Errorf("lagoonId must be unknown in a create preview, got %v", v)
			}
		})
	}
}

// WireDependencies turns off infer's default computedness flow for every output of the
// resource, so each output that is not an input must be wired by hand. This test fails when a
// state type gains an output field. Wire the new field in WireDependencies, add a case to
// previewCases, then add the field name here.
func TestPreviewWiring_CoversEveryOutputField(t *testing.T) {
	cases := []struct {
		name  string
		state any
		want  []string
	}{
		{"DeployTarget", resources.DeployTargetState{}, []string{"created", "lagoonId"}},
		{"Project", resources.ProjectState{}, []string{"created", "lagoonId", "publicKey"}},
		{"Environment", resources.EnvironmentState{}, []string{"created", "lagoonId", "route", "routes"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got []string
			typ := reflect.TypeOf(tc.state)
			for i := 0; i < typ.NumField(); i++ {
				f := typ.Field(i)
				if f.Anonymous {
					continue // embedded inputs mirror the args and keep their value
				}
				got = append(got, strings.Split(f.Tag.Get("pulumi"), ",")[0])
			}
			sort.Strings(got)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("output fields of %s changed: got %v, want %v. Update WireDependencies and previewCases first.", tc.name, got, tc.want)
			}
		})
	}
}
