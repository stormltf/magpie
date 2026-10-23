package gateway

import "testing"

// A route kept with a try marked swapped for Volcengine Ark's endpoint id
// (ep-20261021200314-p76lm) answered under the model behind it
// (deepseek-v4-1-flash-261017) is read as not swapped; another model asked
// for by name stays swapped.
func TestRouteKeptEndpointIDNotSwapped(t *testing.T) {
	r := &Route{Swapped: true, Served: "deepseek-v4-1-flash-261017", Tries: []Try{{Model: "ep-20261021200314-p76lm", Served: "deepseek-v4-1-flash-261017", Swapped: true}}}
	r.routedAgain()
	if r.Swapped || r.Tries[0].Swapped {
		t.Fatalf("still swapped: %+v", r)
	}
	r = &Route{Swapped: true, Served: "gpt-5-mini", Tries: []Try{{Model: "gpt-5", Served: "gpt-5-mini", Swapped: true}}}
	r.routedAgain()
	if !r.Swapped || !r.Tries[0].Swapped {
		t.Fatalf("a real swap was cleared: %+v", r)
	}
}
