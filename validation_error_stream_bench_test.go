// Copyright 2023-2026 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protovalidate

import (
	"strconv"
	"strings"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestValidationErrorErrorEdgeCases(t *testing.T) {
	t.Parallel()
	fieldPath := validate.FieldPath_builder{
		Elements: []*validate.FieldPathElement{
			validate.FieldPathElement_builder{
				FieldName: proto.String("payload"),
			}.Build(),
			validate.FieldPathElement_builder{
				FieldName: proto.String("items"),
				Index:     proto.Uint64(2),
			}.Build(),
			validate.FieldPathElement_builder{
				FieldName: proto.String("value"),
			}.Build(),
		},
	}.Build()
	testCases := []struct {
		name string
		err  *ValidationError
		want string
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			err:  &ValidationError{},
		},
		{
			name: "single nil violation",
			err: &ValidationError{
				Violations: []*Violation{nil},
			},
			want: "validation error: ",
		},
		{
			name: "multiple mixed violations",
			err: &ValidationError{
				Violations: []*Violation{
					nil,
					{},
					{
						Proto: validate.Violation_builder{
							Field:   fieldPath,
							Message: proto.String("message"),
						}.Build(),
					},
					{
						Proto: validate.Violation_builder{
							RuleId: proto.String("rule.id"),
						}.Build(),
					},
				},
			},
			want: "validation errors:\n - \n - [unknown]\n - payload.items[2].value: message\n - [rule.id]",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.want, testCase.err.Error())
		})
	}
}

func TestViolationStringNil(t *testing.T) {
	t.Parallel()

	require.Empty(t, (*Violation)(nil).String())
}

func BenchmarkValidationErrorError(b *testing.B) {
	for _, count := range []int{2, 10, 100} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			err, want := newValidationErrorErrorBenchmarkInput(count)
			if got := err.Error(); got != want {
				b.Fatalf("Error() = %q, want %q", got, want)
			}

			var got string
			for b.Loop() {
				got = err.Error()
			}
			if got != want {
				b.Fatalf("Error() = %q, want %q", got, want)
			}
		})
	}
}

func newValidationErrorErrorBenchmarkInput(count int) (*ValidationError, string) {
	violations := make([]*Violation, count)
	var want strings.Builder
	want.WriteString("validation errors:")
	for i := range violations {
		item := strconv.Itoa(i)
		fieldPath := "payload.items[" + item + "].value"
		message := "is invalid for benchmark item " + item + " with nested field context"
		violations[i] = &Violation{
			Proto: validate.Violation_builder{
				Field: validate.FieldPath_builder{
					Elements: []*validate.FieldPathElement{
						validate.FieldPathElement_builder{
							FieldName: proto.String("payload"),
						}.Build(),
						validate.FieldPathElement_builder{
							FieldName: proto.String("items"),
							Index:     proto.Uint64(uint64(i)),
						}.Build(),
						validate.FieldPathElement_builder{
							FieldName: proto.String("value"),
						}.Build(),
					},
				}.Build(),
				Message: proto.String(message),
			}.Build(),
		}
		want.WriteString("\n - ")
		want.WriteString(fieldPath)
		want.WriteString(": ")
		want.WriteString(message)
	}
	return &ValidationError{Violations: violations}, want.String()
}
