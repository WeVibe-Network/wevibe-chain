package app

import "testing"

func TestCLIEncodingResolvesCustomMsgs(t *testing.T) {
	encCfg := MakeEncodingConfig()

	for _, typeURL := range []string{
		"/wevibe.serve.v1.MsgAnchorPolicyVersion",
		"/wevibe.serve.v1.MsgSubmitEventBatch",
	} {
		if _, err := encCfg.InterfaceRegistry.Resolve(typeURL); err != nil {
			t.Fatalf("expected CLI encoding registry to resolve %s: %v", typeURL, err)
		}
	}
}
