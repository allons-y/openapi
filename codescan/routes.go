// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package codescan

import (
	"fmt"

	spec "github.com/allons-y/openapi-spec"
)

// opConsumesSetter is deprecated in OpenAPI v3 - content types moved to requestBody
// Kept for backward compatibility during annotation parsing
func opConsumesSetter(op *spec.Operation) func([]string) {
	return func(consumes []string) { op.Consumes = consumes }
}

// opProducesSetter is deprecated in OpenAPI v3 - content types moved to response content
// Kept for backward compatibility during annotation parsing
func opProducesSetter(op *spec.Operation) func([]string) {
	return func(produces []string) { op.Produces = produces }
}

// opServersSetter sets operation-level servers for OpenAPI v3
// This replaces the old schemes field which is no longer in v3
func opServersSetter(op *spec.Operation) func([]string) {
	return func(schemes []string) {
		// Convert schemes to servers if needed
		// In v3, operation-level servers override root-level servers
		// For now, we just store as servers with placeholder URLs
		for _, scheme := range schemes {
			op.Servers = append(op.Servers, spec.Server{
				ServerProps: spec.ServerProps{
					URL: scheme + "://",
				},
			})
		}
	}
}

func opSecurityDefsSetter(op *spec.Operation) func([]map[string][]string) {
	return func(securityDefs []map[string][]string) { op.Security = securityDefs }
}

func opResponsesSetter(op *spec.Operation) func(*spec.Response, map[int]spec.Response) {
	return func(def *spec.Response, scr map[int]spec.Response) {
		if op.Responses == nil {
			op.Responses = new(spec.Responses)
		}
		op.Responses.Default = def
		op.Responses.StatusCodeResponses = scr
	}
}

func opParamSetter(op *spec.Operation) func([]*spec.Parameter) {
	return func(params []*spec.Parameter) {
		for _, v := range params {
			op.AddParam(v)
		}
	}
}

func opExtensionsSetter(op *spec.Operation) func(*spec.Extensions) {
	return func(exts *spec.Extensions) {
		for name, value := range *exts {
			op.AddExtension(name, value)
		}
	}
}

type routesBuilder struct {
	ctx         *scanCtx
	route       parsedPathContent
	definitions map[string]spec.Schema
	operations  map[string]*spec.Operation
	responses   map[string]spec.Response
	parameters  []*spec.Parameter
}

func (r *routesBuilder) Build(tgt *spec.Paths) error {
	pthObj := tgt.Paths[r.route.Path]
	op := setPathOperation(
		r.route.Method, r.route.ID,
		&pthObj, r.operations[r.route.ID])

	op.Tags = r.route.Tags

	sp := new(sectionedParser)
	sp.setTitle = func(lines []string) { op.Summary = joinDropLast(lines) }
	sp.setDescription = func(lines []string) { op.Description = joinDropLast(lines) }
	sr := newSetResponses(r.definitions, r.responses, opResponsesSetter(op))
	spa := newSetParams(r.parameters, opParamSetter(op))
	sp.taggers = []tagParser{
		newMultiLineTagParser("Consumes", newMultilineDropEmptyParser(rxConsumes, opConsumesSetter(op)), false),
		newMultiLineTagParser("Produces", newMultilineDropEmptyParser(rxProduces, opProducesSetter(op)), false),
		newSingleLineTagParser("Schemes", newSetSchemes(opServersSetter(op))),
		newMultiLineTagParser("Security", newSetSecurity(rxSecuritySchemes, opSecurityDefsSetter(op)), false),
		newMultiLineTagParser("Parameters", spa, false),
		newMultiLineTagParser("Responses", sr, false),
		newSingleLineTagParser("Deprecated", &setDeprecatedOp{op}),
		newMultiLineTagParser("Extensions", newSetExtensions(opExtensionsSetter(op)), true),
	}
	if err := sp.Parse(r.route.Remaining); err != nil {
		return fmt.Errorf("operation (%s): %w", op.ID, err)
	}

	if tgt.Paths == nil {
		tgt.Paths = make(map[string]spec.PathItem)
	}
	tgt.Paths[r.route.Path] = pthObj
	return nil
}
