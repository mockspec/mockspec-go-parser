package msparser

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

var errInvalidInput = fmt.Errorf("parse: invalid input")
var errObjectMustHaveSingleKey = fmt.Errorf("parse: object must only have one key")
var errUnknownDefinition = fmt.Errorf("parse: unknown definition")

func errInvalidInputReason(reason string) error {
	return fmt.Errorf("%w: %s", errInvalidInput, reason)
}

func errUnknownDefinitionWithName(group, ref string) error {
	return fmt.Errorf("%w: %s (%s)", errUnknownDefinition, ref, group)
}

// ParseFromFile reads the file and returns a Spec object.
func ParseFromFile(name string) (*Spec, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	return Parse(content)
}

// Parse parses the input and returns a Spec object.
// The input must be a valid YAML content.
// The function resolves any references to definitions.
func Parse(input []byte) (*Spec, error) {
	var rSpec any
	if err := yaml.Unmarshal(input, &rSpec); err != nil {
		return nil, err
	}

	rSpecMap, ok := rSpec.(map[string]any)
	if !ok {
		return nil, errInvalidInputReason("spec must be an object")
	}

	spec := createEmpty()

	if err := fillSpec(rSpecMap, spec); err != nil {
		return nil, err
	}

	return spec, nil
}

func createEmpty() *Spec {
	return &Spec{
		Definitions: &Definitions{
			Steps:      map[string][]*Step{},
			Filters:    map[string][]*Filter{},
			Conditions: map[string][]*Condition{},
			Responses:  map[string]*Response{},
		},
		Endpoints: []*Endpoint{},
	}
}

func fillSpec(rSpec map[string]any, spec *Spec) error {
	rSpecDefinitions, err := extractStringMap(rSpec, "definitions")
	if err != nil {
		return err
	}
	if rSpecDefinitions != nil {
		if err := fillDefinitions(rSpecDefinitions, spec.Definitions); err != nil {
			return err
		}
	}

	rSpecEndpoints, err := extractSliceOfAny(rSpec, "endpoints")
	if err != nil {
		return err
	}
	if rSpecEndpoints != nil {
		if err := fillEndpoints(&spec.Endpoints, rSpecEndpoints, spec.Definitions); err != nil {
			return err
		}
	}

	return nil
}

func fillDefinitions(rDefinitions map[string]any, definitions *Definitions) error {
	var err error

	rSteps, err := extractStringMap(rDefinitions, "steps")
	if err != nil {
		return err
	}
	if rSteps != nil {
		if err = fillStepsMap(&definitions.Steps, rSteps, definitions); err != nil {
			return err
		}
	}

	rFilters, err := extractStringMap(rDefinitions, "filters")
	if err != nil {
		return err
	}
	if rFilters != nil {
		if err = fillFiltersMap(&definitions.Filters, rFilters, definitions); err != nil {
			return err
		}
	}

	rConditions, err := extractStringMap(rDefinitions, "conditions")
	if err != nil {
		return err
	}
	if rConditions != nil {
		if err = fillConditionsMap(&definitions.Conditions, rConditions, definitions); err != nil {
			return err
		}
	}

	rResponses, err := extractStringMap(rDefinitions, "responses")
	if err != nil {
		return err
	}
	if rResponses != nil {
		if err = fillResponsesMap(&definitions.Responses, rResponses, definitions); err != nil {
			return err
		}
	}

	return nil
}

func fillStepsMap(stepsMap *map[string][]*Step, rSteps map[string]any, definitions *Definitions) error {
	for stepName, rStepItems := range rSteps {
		rStepItemsSlice, ok := rStepItems.([]any)
		if !ok {
			return errInvalidInputReason(fmt.Sprintf("each step must be an array if step items (%s is not)", stepName))
		}
		var stepsList []*Step
		if err := fillStepsItems(&stepsList, rStepItemsSlice, definitions); err != nil {
			return err
		}
		(*stepsMap)[stepName] = stepsList
	}

	return nil
}

func fillStepsItems(steps *[]*Step, rStepItems []any, definitions *Definitions) error {
	if len(rStepItems) == 0 {
		return errInvalidInputReason("each step must have at least one item")
	}

	for _, rStepItem := range rStepItems {
		rStepItemMap, ok := rStepItem.(map[string]any)
		if !ok {
			return errInvalidInputReason("each step item must be an object")
		}
		if err := fillStepsItem(steps, rStepItemMap, definitions); err != nil {
			return err
		}
	}

	return nil
}

func fillStepsItem(steps *[]*Step, rStepItem map[string]any, definitions *Definitions) error {
	newSteps, err := createSteps(rStepItem, definitions)
	if err != nil {
		return err
	}

	*steps = append(*steps, newSteps...)

	return nil
}

func fillFiltersMap(filtersMap *map[string][]*Filter, rFilters map[string]any, definitions *Definitions) error {
	for filterName, rFilterItems := range rFilters {
		rFilterItemsSlice, ok := rFilterItems.([]any)
		if !ok {
			return errInvalidInputReason(fmt.Sprintf("each filter must be an array of filter items (%s is not)", filterName))
		}
		var filtersList []*Filter
		if err := fillFiltersItems(&filtersList, rFilterItemsSlice, definitions); err != nil {
			return err
		}
		(*filtersMap)[filterName] = filtersList
	}

	return nil
}

func fillFiltersItems(filters *[]*Filter, rFilterItems []any, definitions *Definitions) error {
	if len(rFilterItems) == 0 {
		return errInvalidInputReason("each filter must have at least one item")
	}

	for _, rFilterItem := range rFilterItems {
		rFilterItemMap, ok := rFilterItem.(map[string]any)
		if !ok {
			return errInvalidInputReason("each filter item must be an object")
		}
		if err := fillFiltersItem(filters, rFilterItemMap, definitions); err != nil {
			return err
		}
	}

	return nil
}

func fillFiltersItem(filters *[]*Filter, rFilterItem map[string]any, definitions *Definitions) error {
	newFilters, err := createFilters(definitions, rFilterItem)
	if err != nil {
		return err
	}

	*filters = append(*filters, newFilters...)

	return nil
}

func fillConditionsMap(conditionsMap *map[string][]*Condition, rConditions map[string]any, definitions *Definitions) error {
	for conditionName, rConditionItems := range rConditions {
		rConditionItemsSlice, ok := rConditionItems.([]any)
		if !ok {
			return errInvalidInputReason(fmt.Sprintf("each condition must be an array of condition items (%s is not)", conditionName))
		}
		var conditionsList []*Condition
		if err := fillConditionsItems(&conditionsList, rConditionItemsSlice, definitions); err != nil {
			return err
		}
		(*conditionsMap)[conditionName] = conditionsList
	}

	return nil
}

func fillConditionsItems(conditions *[]*Condition, rConditionItems []any, definitions *Definitions) error {
	if (len(rConditionItems)) == 0 {
		return errInvalidInputReason("each condition must have at least one item")
	}

	for _, rConditionItem := range rConditionItems {
		rConditionItemMap, ok := rConditionItem.(map[string]any)
		if !ok {
			return errInvalidInputReason("each condition item must be an object")
		}
		if err := fillConditionsItem(conditions, rConditionItemMap, definitions); err != nil {
			return err
		}
	}

	return nil
}

func fillConditionsItem(conditions *[]*Condition, rConditionItem map[string]any, definitions *Definitions) error {
	newConditions, err := createConditions(definitions, rConditionItem)
	if err != nil {
		return err
	}

	*conditions = append(*conditions, newConditions...)

	return nil
}

func fillResponsesMap(responsesMap *map[string]*Response, rResponses map[string]any, definitions *Definitions) error {
	for responseName, rResponse := range rResponses {
		rResponseMap, ok := rResponse.(map[string]any)
		if !ok {
			return errInvalidInputReason("each response must be an object")
		}
		var response Response
		if err := fillResponseItem(&response, rResponseMap, definitions); err != nil {
			return err
		}
		(*responsesMap)[responseName] = &response
	}

	return nil
}

func fillResponseItem(response *Response, rResponse map[string]any, definitions *Definitions) error {
	newResponse, err := createResponse(rResponse, definitions)
	if err != nil {
		return err
	}

	*response = *newResponse

	return nil
}

func fillEndpoints(endpoints *[]*Endpoint, rEndpoints []any, definitions *Definitions) error {
	for _, rEndpoint := range rEndpoints {
		rEndpointMap, ok := rEndpoint.(map[string]any)
		if !ok {
			return errInvalidInputReason("each endpoint must be an object")
		}
		endpoint, err := createEndpoint(rEndpointMap, definitions)
		if err != nil {
			return err
		}
		*endpoints = append(*endpoints, endpoint)
	}
	return nil
}

func createSteps(rStepItem map[string]any, definitions *Definitions) ([]*Step, error) {
	var steps []*Step
	var err error

	if len(rStepItem) > 1 {
		return nil, errObjectMustHaveSingleKey
	}

	refName, ok, err := extractString(rStepItem, "$ref")
	if err != nil {
		return nil, err
	}
	if ok {
		return getStepsByReference(definitions, refName)
	}

	for rStepOperation, rStepParams := range rStepItem {
		step := Step{
			Operation: rStepOperation,
		}
		switch rStepParams.(type) {
		case map[string]any:
			step.Parameters = rStepParams.(map[string]any)
		default:
			step.Parameters = map[string]any{
				"value": rStepParams,
			}
		}
		steps = append(steps, &step)
	}

	return steps, nil
}

func createFilters(definitions *Definitions, rFilterItem map[string]any) ([]*Filter, error) {
	refName, ok, err := extractString(rFilterItem, "$ref")
	if err != nil {
		return nil, err
	}
	if ok {
		return getFiltersByReference(definitions, refName)
	}

	filter := Filter{}
	if err = fillStrings(map[string]*string{
		"source": &filter.Source,
		"target": &filter.Target,
	}, rFilterItem); err != nil {
		return nil, err
	}
	if filter.Source == "" {
		return nil, errInvalidInputReason("filter must have a source")
	}

	rSteps, err := extractSliceOfAny(rFilterItem, "steps")
	if err != nil {
		return nil, err
	}
	if rSteps != nil {
		for _, rStep := range rSteps {
			rStepMap, ok := rStep.(map[string]any)
			if !ok {
				return nil, errInvalidInputReason("each step item must be an object")
			}
			steps, err := createSteps(rStepMap, definitions)
			if err != nil {
				return nil, err
			}
			for _, step := range steps {
				filter.Steps = append(filter.Steps, step)
			}
		}
	}

	return []*Filter{&filter}, nil
}

func createConditions(definitions *Definitions, rConditionItem map[string]any) ([]*Condition, error) {
	refName, ok, err := extractString(rConditionItem, "$ref")
	if err != nil {
		return nil, err
	}
	if ok {
		return getConditionsByReference(definitions, refName)
	}

	condition := Condition{}

	sliceAny, err := extractSliceOfAny(rConditionItem, "any")
	if err != nil {
		return nil, err
	}
	if sliceAny != nil {
		if err := fillConditionsItems(&condition.Any, sliceAny, definitions); err != nil {
			return nil, err
		}
		return []*Condition{&condition}, nil
	}

	sliceAll, err := extractSliceOfAny(rConditionItem, "all")
	if err != nil {
		return nil, err
	}
	if sliceAll != nil {
		if err := fillConditionsItems(&condition.All, sliceAll, definitions); err != nil {
			return nil, err
		}
		return []*Condition{&condition}, nil
	}

	if err = fillString(&condition.Source, rConditionItem, "source"); err != nil {
		return nil, err
	}
	if condition.Source == "" {
		return nil, errInvalidInputReason("condition must have a source")
	}

	rChecks, err := extractSliceOfAny(rConditionItem, "checks")
	if err != nil {
		return nil, err
	}
	if rChecks == nil {
		return nil, errInvalidInputReason("checks must be an array")
	}
	condition.Checks, err = createChecks(rChecks)
	if err != nil {
		return nil, err
	}

	return []*Condition{&condition}, nil
}

func createChecks(rChecks []any) ([]*Check, error) {
	var checks []*Check

	for _, rCheck := range rChecks {
		rCheckMap, ok := rCheck.(map[string]any)
		if !ok {
			return nil, errInvalidInputReason("each check item must be an object")
		}
		if len(rCheckMap) != 1 {
			return nil, errObjectMustHaveSingleKey
		}
		for rCheckName, rCheckParams := range rCheckMap {
			check := Check{
				Name: rCheckName,
			}
			switch rCheckParams.(type) {
			case map[string]any:
				check.Parameters = rCheckParams.(map[string]any)
			default:
				check.Parameters = map[string]any{
					"value": rCheckParams,
				}
			}
			checks = append(checks, &check)
		}
	}

	return checks, nil
}

func createResponse(rResponse map[string]any, definitions *Definitions) (*Response, error) {
	var response Response

	refName, ok, err := extractString(rResponse, "$ref")
	if err != nil {
		return nil, err
	}
	if ok {
		return getResponseByReference(definitions, refName)
	}

	if err := fillStrings(map[string]*string{
		"format": &response.Format,
		"body":   &response.Body,
	}, rResponse); err != nil {
		return nil, err
	}

	if err := fillInt(&response.Status, rResponse, "status"); err != nil {
		return nil, err
	}

	rHeaders, err := extractStringMap(rResponse, "headers")
	if err != nil {
		return nil, err
	}
	if rHeaders != nil {
		headers := map[string][]string{}
		for headerName, headerValues := range rHeaders {
			headers[headerName] = []string{}

			headerValuesSlice, ok := headerValues.([]any)
			if !ok {
				return nil, errInvalidInputReason("each header values must be an array of strings")
			}

			for _, headerValue := range headerValuesSlice {
				headers[headerName] = append(headers[headerName], headerValue.(string))
			}
		}
	}

	return &response, nil
}

func createEndpoint(rEndpoint map[string]any, definitions *Definitions) (*Endpoint, error) {
	var endpoint Endpoint

	if err := fillStrings(map[string]*string{
		"description": &endpoint.Description,
		"host":        &endpoint.Host,
		"method":      &endpoint.Method,
		"path":        &endpoint.Path,
		"bodyFormat":  &endpoint.BodyFormat,
	}, rEndpoint); err != nil {
		return nil, err
	}

	rFilterItems, err := extractSliceOfAny(rEndpoint, "filters")
	if err != nil {
		return nil, err
	}
	if rFilterItems != nil {
		if err := fillFiltersItems(&endpoint.Filters, rFilterItems, definitions); err != nil {
			return nil, err
		}
	}

	rConditionItems, err := extractSliceOfAny(rEndpoint, "conditions")
	if err != nil {
		return nil, err
	}
	if rConditionItems != nil {
		if err := fillConditionsItems(&endpoint.Conditions, rConditionItems, definitions); err != nil {
			return nil, err
		}
	}

	rEndpointItems, err := extractSliceOfAny(rEndpoint, "endpoints")
	if err != nil {
		return nil, err
	}
	if rEndpointItems != nil {
		if err := fillEndpoints(&endpoint.Endpoints, rEndpointItems, definitions); err != nil {
			return nil, err
		}
	}

	rResponse, err := extractStringMap(rEndpoint, "response")
	if err != nil {
		return nil, err
	}
	if rResponse != nil {
		var response Response
		if err := fillResponseItem(&response, rResponse, definitions); err != nil {
			return nil, err
		}
		endpoint.Response = &response
	}

	if len(endpoint.Endpoints) == 0 && endpoint.Response == nil {
		return nil, errInvalidInputReason("endpoint must have either sub-endpoints or a response")
	}

	return &endpoint, nil
}

func getStepsByReference(definitions *Definitions, refName string) ([]*Step, error) {
	steps, ok := definitions.Steps[refName]
	if !ok {
		return nil, errUnknownDefinitionWithName("steps", refName)
	}

	return steps, nil
}

func getFiltersByReference(definitions *Definitions, refName string) ([]*Filter, error) {
	filters, ok := definitions.Filters[refName]
	if !ok {
		return nil, errUnknownDefinitionWithName("filters", refName)
	}

	return filters, nil
}

func getConditionsByReference(definitions *Definitions, refName string) ([]*Condition, error) {
	conditions, ok := definitions.Conditions[refName]
	if !ok {
		return nil, errUnknownDefinitionWithName("conditions", refName)
	}

	return conditions, nil
}

func getResponseByReference(definitions *Definitions, refName string) (*Response, error) {
	response, ok := definitions.Responses[refName]
	if !ok {
		return nil, errUnknownDefinitionWithName("responses", refName)
	}

	return response, nil
}

func extractString(input map[string]any, name string) (string, bool, error) {
	value, ok := input[name]
	if !ok {
		return "", false, nil
	}

	str, ok := value.(string)
	if !ok {
		return "", false, errInvalidInputReason(fmt.Sprintf("expected '%s' to be a string", name))
	}

	return str, true, nil
}

func extractInt(input map[string]any, name string) (int, bool, error) {
	value, ok := input[name]
	if !ok {
		return 0, false, nil
	}

	i, ok := value.(int)
	if !ok {
		return 0, false, errInvalidInputReason(fmt.Sprintf("expect '%s' to be an integer", name))
	}

	return i, true, nil
}

func fillString(target *string, input map[string]any, name string) error {
	value, ok, err := extractString(input, name)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	*target = value

	return nil
}

func fillStrings(targets map[string]*string, input map[string]any) error {
	for name, target := range targets {
		if err := fillString(target, input, name); err != nil {
			return err
		}
	}
	return nil
}

func fillInt(target *int, input map[string]any, name string) error {
	value, ok, err := extractInt(input, name)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	*target = value

	return nil
}

func extractStringMap(input map[string]any, name string) (map[string]any, error) {
	output, ok := input[name]
	if !ok {
		return nil, nil
	}

	outputMap, ok := output.(map[string]any)
	if !ok {
		return nil, errInvalidInputReason(fmt.Sprintf("expected '%s' to be an object", name))
	}

	return outputMap, nil
}

func extractSliceOfAny(input map[string]any, name string) ([]any, error) {
	output, ok := input[name]
	if !ok {
		return nil, nil
	}

	outputSlice, ok := output.([]any)
	if !ok {
		return nil, errInvalidInputReason(fmt.Sprintf("expected '%s' to be an array", name))
	}

	return outputSlice, nil
}
