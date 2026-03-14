package grpc_test

import (
	"testing"

	"github.com/DriftaBot/driftabot-engine/internal/classifier"
	differgrpc "github.com/DriftaBot/driftabot-engine/internal/differ/grpc"
	parsergrpc "github.com/DriftaBot/driftabot-engine/internal/parser/grpc"
	"github.com/DriftaBot/driftabot-engine/pkg/schema"
)

const testdataDir = "../../testdata/"

// --------------------------------------------------------------------------
// Severity rules — table-driven
// --------------------------------------------------------------------------

type severityCase struct {
	name     string
	change   schema.Change
	expected schema.Severity
}

var severityCases = []severityCase{
	// Service-level
	{
		name:     "service removed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCServiceRemoved, Before: "UserService"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "service added is non-breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCServiceAdded, After: "AdminService"},
		expected: schema.SeverityNonBreaking,
	},

	// RPC-level
	{
		name:     "rpc removed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCRPCRemoved, Before: "GetUser"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "rpc added is non-breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCRPCAdded, After: "CreateUser"},
		expected: schema.SeverityNonBreaking,
	},
	{
		name:     "rpc request type changed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCRPCRequestTypeChanged, Before: "GetUserRequest", After: "GetUserRequestV2"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "rpc response type changed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCRPCResponseTypeChanged, Before: "GetUserResponse", After: "GetUserResponseV2"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "rpc streaming mode changed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCRPCStreamingChanged, Before: "unary", After: "server streaming"},
		expected: schema.SeverityBreaking,
	},

	// Message-level
	{
		name:     "message removed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCMessageRemoved, Before: "GetUserRequest"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "message added is non-breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCMessageAdded, After: "NewMessage"},
		expected: schema.SeverityNonBreaking,
	},

	// Field-level
	{
		name:     "field removed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCFieldRemoved, Before: "string"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "field added is non-breaking (proto3 optional by default)",
		change:   schema.Change{Type: schema.ChangeTypeGRPCFieldAdded, After: "string"},
		expected: schema.SeverityNonBreaking,
	},
	{
		name:     "field type changed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCFieldTypeChanged, Before: "string", After: "bytes"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "field number changed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCFieldNumberChanged, Before: "1", After: "2"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "field label singular to repeated is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCFieldLabelChanged, Before: "singular", After: "repeated"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "field label repeated to singular is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGRPCFieldLabelChanged, Before: "repeated", After: "singular"},
		expected: schema.SeverityBreaking,
	},
}

func TestClassify_GRPC_SeverityRules(t *testing.T) {
	for _, tc := range severityCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifier.Classify("base", "head", []schema.Change{tc.change})
			if len(result.Changes) != 1 {
				t.Fatalf("expected 1 classified change, got %d", len(result.Changes))
			}
			got := result.Changes[0].Severity
			if got != tc.expected {
				t.Errorf("severity = %s, want %s", got, tc.expected)
			}
		})
	}
}

// --------------------------------------------------------------------------
// Summary counts
// --------------------------------------------------------------------------

func TestClassify_GRPC_SummaryCounts(t *testing.T) {
	changes := []schema.Change{
		{Type: schema.ChangeTypeGRPCRPCRemoved},       // breaking
		{Type: schema.ChangeTypeGRPCFieldRemoved},     // breaking
		{Type: schema.ChangeTypeGRPCServiceAdded},     // non-breaking
		{Type: schema.ChangeTypeGRPCFieldAdded},       // non-breaking
	}

	result := classifier.Classify("base", "head", changes)

	if result.Summary.Total != 4 {
		t.Errorf("total = %d, want 4", result.Summary.Total)
	}
	if result.Summary.Breaking != 2 {
		t.Errorf("breaking = %d, want 2", result.Summary.Breaking)
	}
	if result.Summary.NonBreaking != 2 {
		t.Errorf("non_breaking = %d, want 2", result.Summary.NonBreaking)
	}
}

// --------------------------------------------------------------------------
// Integration: fixture files → full pipeline
// --------------------------------------------------------------------------

func TestClassify_GRPC_FixtureBreakingChanges(t *testing.T) {
	base, err := parsergrpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	head, err := parsergrpc.Parse(testdataDir + "head.proto")
	if err != nil {
		t.Fatalf("parse head: %v", err)
	}

	changes := differgrpc.Diff(base, head)
	result := classifier.Classify("base.proto", "head.proto", changes)

	if result.Summary.Breaking == 0 {
		t.Error("expected at least one breaking change from fixture diff")
	}
}

func TestClassify_GRPC_FixtureNonBreakingChanges(t *testing.T) {
	base, err := parsergrpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	head, err := parsergrpc.Parse(testdataDir + "head.proto")
	if err != nil {
		t.Fatalf("parse head: %v", err)
	}

	changes := differgrpc.Diff(base, head)
	result := classifier.Classify("base.proto", "head.proto", changes)

	if result.Summary.NonBreaking == 0 {
		t.Error("expected at least one non-breaking change from fixture diff")
	}
}

func TestClassify_GRPC_IdenticalSchemas_NoChanges(t *testing.T) {
	base, err := parsergrpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	result := classifier.Classify("base.proto", "base.proto", differgrpc.Diff(base, base))

	if result.Summary.Total != 0 {
		t.Errorf("expected 0 changes for identical schemas, got %d", result.Summary.Total)
	}
}
