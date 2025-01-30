package msparser

// Spec is a root object that contains all the information about the mock.
type Spec struct {
	Definitions *Definitions
	Endpoints   []*Endpoint
}

// Definitions contains all the definitions for the mock.
type Definitions struct {
	Steps      map[string][]*Step
	Filters    map[string][]*Filter
	Conditions map[string][]*Condition
	Responses  map[string]*Response
}

// Endpoint defines rules for a specific endpoint that, when matched, will return a specific response.
// It can also contain sub-endpoints which will be checked if the parent endpoint is matched.
type Endpoint struct {
	Description string
	Host        string
	Method      string
	Path        string
	BodyFormat  string
	Filters     []*Filter
	Conditions  []*Condition
	Endpoints   []*Endpoint
	Response    *Response
}

// Step is a single operation which is a part of a Filter.
type Step struct {
	Operation  string
	Parameters map[string]any
}

// Filter is a set of [Step]s that can be applied to a parameter value.
// Parameters can come from the endpoint path, query, headers, or body.
//
// If [Filter.Target] is not empty, the filter will be set as the [Filter.Target] parameter.
// If [Filter.Target] is empty, the filter will be applied to the [Filter.Source] parameter, but always
// in parameters even if the original value was taken from query, headers, or body.
type Filter struct {
	Source string
	Target string
	Steps  []*Step
}

// Condition is a set of [Check]s that must be satisfied for the [Condition] to be true.
// If the [Condition] is true, the endpoint is matched.
//
// If [Condition.Any] or [Condition.All] is not empty, the condition will be true if any or all of their conditions are true.
//
// Only one of [Condition.Any], [Condition.All], or [Condition.Source] with [Condition.Checks] can be set.
type Condition struct {
	Any    []*Condition
	All    []*Condition
	Source string
	Checks []*Check
}

// Check is a single operation that must be satisfied for the [Condition] to be true.
type Check struct {
	Name       string
	Parameters map[string]any
}

// Response defines the response that will be returned when the endpoint is matched.
//
// [Response.Format] can be one of: raw, json, or xml. If other than raw, the Content-Type header will be set accordingly.
type Response struct {
	Status  int
	Format  string
	Headers map[string][]string
	Body    string
}
