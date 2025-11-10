package services

import (
	"strings"
	"testing"
)

// TestHTMLFragmentToSSE tests SSE format conversion
func TestHTMLFragmentToSSE(t *testing.T) {
	t.Run("ConvertFragmentWithTarget", func(t *testing.T) {
		fragment := HTMLFragmentEvent{
			Type:       "test_event",
			Target:     "#my-target",
			SwapMethod: "innerHTML",
			HTML:       "<div>Test Content</div>",
		}

		sseString := HTMLFragmentToSSE(fragment)

		// Should contain event type
		AssertTrue(t, strings.Contains(sseString, "event: test_event"), "Should contain event type")

		// Should contain data field with JSON
		AssertTrue(t, strings.Contains(sseString, "data:"), "Should contain data field")
		AssertTrue(t, strings.Contains(sseString, "\"target\":\"#my-target\""), "Should contain target")
		AssertTrue(t, strings.Contains(sseString, "\"swap\":\"innerHTML\""), "Should contain swap method")
		AssertTrue(t, strings.Contains(sseString, "\"html\""), "Should contain html field")

		// Should end with double newline (SSE format)
		AssertTrue(t, strings.HasSuffix(sseString, "\n\n"), "Should end with double newline")
	})

	t.Run("ConvertFragmentWithoutTarget", func(t *testing.T) {
		fragment := HTMLFragmentEvent{
			Type:       "simple_event",
			Target:     "",
			SwapMethod: "",
			HTML:       "<div>Simple HTML</div>",
		}

		sseString := HTMLFragmentToSSE(fragment)

		// Should contain event type
		AssertTrue(t, strings.Contains(sseString, "event: simple_event"), "Should contain event type")

		// Should contain HTML
		AssertTrue(t, strings.Contains(sseString, "data:"), "Should contain data field")
		AssertTrue(t, strings.Contains(sseString, "Simple HTML"), "Should contain HTML content")
	})

	t.Run("HTMLFragmentEventStructure", func(t *testing.T) {
		// Test that HTMLFragmentEvent struct can be created
		fragment := HTMLFragmentEvent{
			Type:       "join_request",
			Target:     "#join-requests",
			SwapMethod: "beforeend",
			HTML:       "<div>New Request</div>",
		}

		AssertEqual(t, "join_request", fragment.Type, "Type should be set")
		AssertEqual(t, "#join-requests", fragment.Target, "Target should be set")
		AssertEqual(t, "beforeend", fragment.SwapMethod, "SwapMethod should be set")
		AssertTrue(t, len(fragment.HTML) > 0, "HTML should not be empty")
	})
}

// TestHTMLFragmentEvent_AllSwapMethods tests different HTMX swap methods
func TestHTMLFragmentEvent_AllSwapMethods(t *testing.T) {
	swapMethods := []string{"innerHTML", "outerHTML", "beforebegin", "afterbegin", "beforeend", "afterend"}

	for _, method := range swapMethods {
		t.Run("SwapMethod_"+method, func(t *testing.T) {
			fragment := HTMLFragmentEvent{
				Type:       "test",
				Target:     "#target",
				SwapMethod: method,
				HTML:       "<div>Test</div>",
			}

			sseString := HTMLFragmentToSSE(fragment)
			AssertTrue(t, strings.Contains(sseString, "\"swap\":\""+method+"\""),
				"SSE should contain swap method: "+method)
		})
	}
}
