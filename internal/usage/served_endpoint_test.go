package usage

import "testing"

// Volcengine Ark asks for a model by its inference endpoint's id
// (ep-20261021200314-p76lm) and answers with the model behind it
// (deepseek-v4-1-flash-261017): an endpoint id names no model, so there is
// nothing to compare the reply's name with. A model asked for by name is
// still judged by it.
func TestSwappedArkEndpointID(t *testing.T) {
	for _, c := range [][2]string{
		{"ep-20261021200314-p76lm", "deepseek-v4-1-flash-261017"},
		{"volcengine/ep-20261021200314-p76lm", "deepseek-v4-1-flash-261017"},
		{"ep-20261021200314-p76lm", "deepseek-v4-1-flash"},
		{"EP-20261021200314-P76LM", "deepseek-v4-1-flash-261017"},
		{"ep-20240918abcd-p76lm", "doubao-seed-1-6-250615"},
	} {
		if Swapped(c[0], c[1]) {
			t.Errorf("%s answered as %s is the model the endpoint serves", c[0], c[1])
		}
	}
	for _, c := range [][2]string{
		{"deepseek-v4.1-flash", "deepseek-v4.1-pro"},
		{"my-ep-20261021200314-p76lm", "deepseek-v4-1-flash-261017"},
	} {
		if !Swapped(c[0], c[1]) {
			t.Errorf("%s answered as %s is another model", c[0], c[1])
		}
	}
}

// A row kept marked swapped before an endpoint id was read as its
// deployment's own name is read back as not swapped.
func TestKeptRowArkEndpointIDNotSwapped(t *testing.T) {
	c := &rowChunk{}
	c.add(Row{Record: Record{Model: "ep-20261021200314-p76lm", Served: "deepseek-v4-1-flash-261017"}, Swapped: true}, "", 0, false)
	c.add(Row{Record: Record{Model: "deepseek-v4.1-flash", Served: "deepseek-v4.1-pro"}, Swapped: true}, "", 1, false)
	for _, chunk := range []*rowChunk{c, c.pack().unpack()} {
		if chunk.row(0).Swapped || !chunk.row(1).Swapped {
			t.Fatalf("endpoint %v, other model %v", chunk.row(0).Swapped, chunk.row(1).Swapped)
		}
	}
}
