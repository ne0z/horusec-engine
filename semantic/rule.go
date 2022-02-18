// Copyright 2020 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package semantic

import (
	"fmt"
	"os"

	"github.com/ZupIT/horusec-devkit/pkg/enums/languages"
	engine "github.com/ZupIT/horusec-engine"
	"github.com/ZupIT/horusec-engine/internal/ast"
	javascript "github.com/ZupIT/horusec-engine/internal/horusec-javascript"
	"github.com/ZupIT/horusec-engine/internal/ir"
	"github.com/ZupIT/horusec-engine/semantic/analysis"
)

// Rule is a generic implementation of engine.Rule interface that evaluate the analysis
// using the semantic engine API.
//
// Different from text.Rule implementation, Rule will aplly the rules using the semantic
// of the language, instead of applying regex on raw source code. Rule basically wraps
// the entire process of parsing the value and running the analyzer on this parsed file.
type Rule struct {
	engine.Metadata
	Language languages.Language // Language that will be parsed.
	Analyzer analysis.Analyzer  // Analyzer entrypoint to be used on analysis.
}

// Run implements engine.Rule.Run.
func (r *Rule) Run(path string) ([]engine.Finding, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var ast *ast.File
	switch r.Language {
	case languages.Javascript:
		ast, err = javascript.ParseFile(path, src)
	default:
		return nil, fmt.Errorf("language %s not supported", r.Language)
	}
	if err != nil {
		return nil, fmt.Errorf("parse %s file: %w", r.Language, err)
	}

	f := ir.NewFile(ast)
	f.Build()

	var findings []engine.Finding

	report := func(issue analysis.Issue) {
		findings = append(findings, engine.Finding{
			ID:          r.ID,
			Name:        r.Name,
			Severity:    r.Severity,
			Confidence:  r.Confidence,
			Description: r.Description,
			CodeSample:  string(src[issue.StartOffset:issue.EndOffset]),
			SourceLocation: engine.Location{
				Filename: path,
				Line:     issue.Line,
				Column:   issue.Column,
			},
		})
	}

	for _, member := range f.Members {
		if fn, ok := member.(*ir.Function); ok {
			r.Analyzer.Run(&analysis.Pass{
				File:     f,
				Function: fn,
				Report:   report,
			})
		}
	}

	return findings, nil
}