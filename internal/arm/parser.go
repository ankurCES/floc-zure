// Package arm implements ARM template parsing, parameter resolution,
// variable expansion, and basic ARM template function evaluation.
package arm

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Template represents a parsed ARM template.
type Template struct {
	Schema         string                        `json:"$schema"`
	ContentVersion string                        `json:"contentVersion"`
	Parameters     map[string]ParameterDef       `json:"parameters,omitempty"`
	Variables      map[string]interface{}         `json:"variables,omitempty"`
	Resources      []ResourceDef                 `json:"resources"`
	Outputs        map[string]OutputDef           `json:"outputs,omitempty"`
}

// ParameterDef defines an ARM template parameter.
type ParameterDef struct {
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
	AllowedValues []interface{} `json:"allowedValues,omitempty"`
	MinLength    *int        `json:"minLength,omitempty"`
	MaxLength    *int        `json:"maxLength,omitempty"`
	Metadata     *ParamMeta  `json:"metadata,omitempty"`
}

// ParamMeta holds parameter metadata (description).
type ParamMeta struct {
	Description string `json:"description,omitempty"`
}

// ResourceDef defines a resource in an ARM template.
type ResourceDef struct {
	Type       string                 `json:"type"`
	APIVersion string                `json:"apiVersion"`
	Name       string                 `json:"name"`
	Location   string                 `json:"location"`
	DependsOn  []string               `json:"dependsOn,omitempty"`
	Tags       map[string]string      `json:"tags,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	SKU        map[string]interface{} `json:"sku,omitempty"`
	Kind       string                 `json:"kind,omitempty"`
}

// OutputDef defines an ARM template output.
type OutputDef struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// ParameterValues maps parameter names to their supplied values.
type ParameterValues struct {
	Schema         string                          `json:"$schema,omitempty"`
	ContentVersion string                          `json:"contentVersion,omitempty"`
	Parameters     map[string]ParameterValueEntry  `json:"parameters"`
}

// ParameterValueEntry wraps a parameter value in the standard ARM format.
type ParameterValueEntry struct {
	Value interface{} `json:"value"`
}

// ParseResult holds the fully resolved template ready for deployment.
type ParseResult struct {
	Template   *Template
	Parameters map[string]interface{} // resolved parameter values
	Variables  map[string]interface{} // resolved variable values
	Resources  []ResolvedResource     // resources with all expressions evaluated
}

// ResolvedResource is a resource with all ARM expressions resolved to concrete values.
type ResolvedResource struct {
	Type       string                 `json:"type"`
	APIVersion string                `json:"apiVersion"`
	Name       string                 `json:"name"`
	Location   string                 `json:"location"`
	DependsOn  []string               `json:"dependsOn,omitempty"`
	Tags       map[string]string      `json:"tags,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	SKU        map[string]interface{} `json:"sku,omitempty"`
	Kind       string                 `json:"kind,omitempty"`
}

// ParseTemplate reads and parses an ARM template JSON file.
func ParseTemplate(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}
	return ParseTemplateBytes(data)
}

// ParseTemplateBytes parses ARM template JSON from bytes.
func ParseTemplateBytes(data []byte) (*Template, error) {
	var t Template
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parse template JSON: %w", err)
	}
	if len(t.Resources) == 0 {
		return nil, fmt.Errorf("template has no resources")
	}
	return &t, nil
}

// ParseParameterFile reads a parameters JSON file.
func ParseParameterFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read parameters file: %w", err)
	}
	return ParseParameterBytes(data)
}

// ParseParameterBytes parses parameter values from JSON bytes.
func ParseParameterBytes(data []byte) (map[string]interface{}, error) {
	var pv ParameterValues
	if err := json.Unmarshal(data, &pv); err != nil {
		return nil, fmt.Errorf("parse parameters JSON: %w", err)
	}
	result := make(map[string]interface{}, len(pv.Parameters))
	for k, v := range pv.Parameters {
		result[k] = v.Value
	}
	return result, nil
}

// Resolve merges supplied parameter values with defaults, evaluates variables,
// and resolves all ARM expressions in resource definitions.
func Resolve(t *Template, suppliedParams map[string]interface{}) (*ParseResult, error) {
	// 1. Resolve parameters: supplied values override defaults
	params := make(map[string]interface{}, len(t.Parameters))
	for name, def := range t.Parameters {
		if val, ok := suppliedParams[name]; ok {
			params[name] = val
		} else if def.DefaultValue != nil {
			params[name] = def.DefaultValue
		} else {
			return nil, fmt.Errorf("missing required parameter %q", name)
		}
	}

	// 2. Build evaluation context
	ctx := &evalContext{
		parameters: params,
		variables:  make(map[string]interface{}),
	}

	// 3. Resolve variables (may reference parameters)
	for name, val := range t.Variables {
		ctx.variables[name] = ctx.resolveValue(val)
	}

	// 4. Resolve each resource
	resolved := make([]ResolvedResource, 0, len(t.Resources))
	for _, r := range t.Resources {
		rr := ResolvedResource{
			Type:       ctx.resolveString(r.Type),
			APIVersion: r.APIVersion,
			Name:       ctx.resolveString(r.Name),
			Location:   ctx.resolveString(r.Location),
			DependsOn:  r.DependsOn,
			Kind:       ctx.resolveString(r.Kind),
		}
		if r.Tags != nil {
			rr.Tags = make(map[string]string, len(r.Tags))
			for k, v := range r.Tags {
				rr.Tags[k] = ctx.resolveString(v)
			}
		}
		if r.Properties != nil {
			rr.Properties = ctx.resolveMap(r.Properties)
		}
		if r.SKU != nil {
			rr.SKU = ctx.resolveMap(r.SKU)
		}
		resolved = append(resolved, rr)
	}

	return &ParseResult{
		Template:   t,
		Parameters: params,
		Variables:  ctx.variables,
		Resources:  resolved,
	}, nil
}

// evalContext holds parameter/variable values for expression evaluation.
type evalContext struct {
	parameters map[string]interface{}
	variables  map[string]interface{}
}

// ARM expression pattern: "[functionName('arg1', 'arg2')]" or "[concat(...)]"
var exprPattern = regexp.MustCompile(`^\[(.+)\]$`)

// resolveString evaluates ARM expressions in a string value.
func (c *evalContext) resolveString(s string) string {
	if s == "" {
		return s
	}
	// Check if the entire string is an ARM expression
	if m := exprPattern.FindStringSubmatch(s); m != nil {
		result := c.evalExpr(m[1])
		if str, ok := result.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", result)
	}
	return s
}

// resolveValue recursively resolves ARM expressions in any value.
func (c *evalContext) resolveValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		if m := exprPattern.FindStringSubmatch(val); m != nil {
			return c.evalExpr(m[1])
		}
		return val
	case map[string]interface{}:
		return c.resolveMap(val)
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, item := range val {
			out[i] = c.resolveValue(item)
		}
		return out
	default:
		return v
	}
}

// resolveMap resolves all values in a map.
func (c *evalContext) resolveMap(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = c.resolveValue(v)
	}
	return out
}

// evalExpr evaluates an ARM template expression (the content between [ and ]).
// Supports: parameters(), variables(), concat(), resourceGroup().location,
// uniqueString(), toLower(), toUpper(), resourceId(), format().
func (c *evalContext) evalExpr(expr string) interface{} {
	expr = strings.TrimSpace(expr)

	// parameters('name')
	if strings.HasPrefix(expr, "parameters(") {
		name := extractSingleArg(expr, "parameters")
		if val, ok := c.parameters[name]; ok {
			return val
		}
		return expr
	}

	// variables('name')
	if strings.HasPrefix(expr, "variables(") {
		name := extractSingleArg(expr, "variables")
		if val, ok := c.variables[name]; ok {
			return val
		}
		return expr
	}

	// concat(a, b, c, ...)
	if strings.HasPrefix(expr, "concat(") {
		args := extractFuncArgs(expr, "concat")
		var sb strings.Builder
		for _, arg := range args {
			sb.WriteString(fmt.Sprintf("%v", c.evalExpr(arg)))
		}
		return sb.String()
	}

	// toLower(expr)
	if strings.HasPrefix(expr, "toLower(") {
		inner := extractInnerExpr(expr, "toLower")
		return strings.ToLower(fmt.Sprintf("%v", c.evalExpr(inner)))
	}

	// toUpper(expr)
	if strings.HasPrefix(expr, "toUpper(") {
		inner := extractInnerExpr(expr, "toUpper")
		return strings.ToUpper(fmt.Sprintf("%v", c.evalExpr(inner)))
	}

	// resourceGroup().location
	if strings.HasPrefix(expr, "resourceGroup()") {
		if strings.HasSuffix(expr, ".location") {
			if loc, ok := c.parameters["location"]; ok {
				return loc
			}
			return "eastus"
		}
		return "resourceGroup()"
	}

	// uniqueString(args...) — deterministic hash for simulation
	if strings.HasPrefix(expr, "uniqueString(") {
		args := extractFuncArgs(expr, "uniqueString")
		var parts []string
		for _, a := range args {
			parts = append(parts, fmt.Sprintf("%v", c.evalExpr(a)))
		}
		return deterministicHash(strings.Join(parts, "-"))
	}

	// resourceId(type, name) — build a fake resource ID
	if strings.HasPrefix(expr, "resourceId(") {
		args := extractFuncArgs(expr, "resourceId")
		if len(args) >= 2 {
			rType := fmt.Sprintf("%v", c.evalExpr(args[0]))
			rName := fmt.Sprintf("%v", c.evalExpr(args[1]))
			return fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/arm-deployed/providers/%s/%s", rType, rName)
		}
		return expr
	}

	// format('fmt', args...) — simple string formatting
	if strings.HasPrefix(expr, "format(") {
		args := extractFuncArgs(expr, "format")
		if len(args) >= 1 {
			fmtStr := fmt.Sprintf("%v", c.evalExpr(args[0]))
			for i := 1; i < len(args); i++ {
				placeholder := fmt.Sprintf("{%d}", i-1)
				fmtStr = strings.ReplaceAll(fmtStr, placeholder, fmt.Sprintf("%v", c.evalExpr(args[i])))
			}
			return fmtStr
		}
		return expr
	}

	// String literal: 'value'
	if strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'") {
		return expr[1 : len(expr)-1]
	}

	return expr
}

// extractSingleArg extracts the single-quoted argument from func('arg').
func extractSingleArg(expr, funcName string) string {
	// funcName('arg')
	inner := expr[len(funcName)+1 : len(expr)-1] // strip funcName( and )
	inner = strings.TrimSpace(inner)
	if strings.HasPrefix(inner, "'") && strings.HasSuffix(inner, "'") {
		return inner[1 : len(inner)-1]
	}
	return inner
}

// extractInnerExpr extracts the inner expression from func(expr).
func extractInnerExpr(expr, funcName string) string {
	return strings.TrimSpace(expr[len(funcName)+1 : len(expr)-1])
}

// extractFuncArgs splits function arguments respecting nested parentheses and quotes.
func extractFuncArgs(expr, funcName string) []string {
	inner := expr[len(funcName)+1 : len(expr)-1]
	var args []string
	depth := 0
	inQuote := false
	start := 0
	for i := 0; i < len(inner); i++ {
		ch := inner[i]
		if ch == '\'' {
			inQuote = !inQuote
		} else if !inQuote {
			if ch == '(' {
				depth++
			} else if ch == ')' {
				depth--
			} else if ch == ',' && depth == 0 {
				args = append(args, strings.TrimSpace(inner[start:i]))
				start = i + 1
			}
		}
	}
	args = append(args, strings.TrimSpace(inner[start:]))
	return args
}

// deterministicHash produces a short deterministic string from input (simulates uniqueString).
func deterministicHash(input string) string {
	h := uint32(0)
	for _, c := range input {
		h = h*31 + uint32(c)
	}
	return fmt.Sprintf("%08x%05x", h, h>>16)
}
