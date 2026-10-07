package clashapi

import (
	"encoding/json"
	"sync/atomic"
	"testing"

	"github.com/sagernet/sing-box/adapter"
	"github.com/stretchr/testify/require"
)

type payloadTestRule struct{ adapter.Rule }

func (payloadTestRule) String() string             { return "domain_suffix=[dell.com 2ip.io vencord.dev...]" }
func (payloadTestRule) Action() adapter.RuleAction { return nil }

func TestXConnectionPayload(t *testing.T) {
	c := connectionObject{
		Metadata: adapter.InboundContext{Domain: "chatgpt.com", RuleMatchPayload: "domain_suffix=chatgpt.com"},
		Upload:   new(atomic.Int64), Download: new(atomic.Int64), Rule: payloadTestRule{},
	}
	encoded, err := c.MarshalJSON()
	require.NoError(t, err)
	var response map[string]any
	require.NoError(t, json.Unmarshal(encoded, &response))
	require.Equal(t, "domain_suffix=chatgpt.com", response["rulePayload"])
	c.Rule = nil
	encoded, err = c.MarshalJSON()
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(encoded, &response))
	require.Equal(t, "final", response["rule"])
	require.Equal(t, "", response["rulePayload"])
}
