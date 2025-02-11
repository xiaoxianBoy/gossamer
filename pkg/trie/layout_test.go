// Copyright 2022 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package trie

import (
	"testing"

	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/stretchr/testify/assert"
)

func Test_Version_String(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		version       TrieLayout
		versionString string
		panicMessage  string
	}{
		"v0": {
			version:       V0,
			versionString: "v0",
		},
		"invalid": {
			version:      TrieLayout(99),
			panicMessage: "unknown version 99",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if testCase.panicMessage != "" {
				assert.PanicsWithValue(t, testCase.panicMessage, func() {
					_ = testCase.version.String()
				})
				return
			}

			versionString := testCase.version.String()
			assert.Equal(t, testCase.versionString, versionString)
		})
	}
}

func Test_ParseVersion(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		v          any
		version    TrieLayout
		errWrapped error
		errMessage string
	}{
		"v0": {
			v:       "v0",
			version: V0,
		},
		"V0": {
			v:       "V0",
			version: V0,
		},
		"0": {
			v:       uint8(0),
			version: V0,
		},
		"v1": {
			v:       "v1",
			version: V1,
		},
		"V1": {
			v:       "V1",
			version: V1,
		},
		"1": {
			v:       uint8(1),
			version: V1,
		},
		"invalid": {
			v:          "xyz",
			errWrapped: ErrParseVersion,
			errMessage: "parsing version failed: \"xyz\" must be one of [v0, v1]",
		},
		"invalid_uint8": {
			v:          uint8(99),
			errWrapped: ErrParseVersion,
			errMessage: "parsing version failed: \"V99\" must be one of [v0, v1]",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var version TrieLayout

			var err error
			switch typed := testCase.v.(type) {
			case string:
				version, err = ParseVersion(typed)
			case uint8:
				version, err = ParseVersion(typed)
			default:
				t.Fail()
			}

			assert.Equal(t, testCase.version, version)
			assert.ErrorIs(t, err, testCase.errWrapped)
			if testCase.errWrapped != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
		})
	}
}

func Test_Version_MaxInlineValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		version      TrieLayout
		max          int
		panicMessage string
	}{
		"v0": {
			version: V0,
			max:     NoMaxInlineValueSize,
		},
		"v1": {
			version: V1,
			max:     V1MaxInlineValueSize,
		},
		"invalid": {
			version:      TrieLayout(99),
			max:          0,
			panicMessage: "unknown version 99",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if testCase.panicMessage != "" {
				assert.PanicsWithValue(t, testCase.panicMessage, func() {
					_ = testCase.version.MaxInlineValue()
				})
				return
			}

			maxInline := testCase.version.MaxInlineValue()
			assert.Equal(t, testCase.max, maxInline)
		})
	}
}

func Test_Version_Root(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		version  TrieLayout
		input    Entries
		expected common.Hash
	}{
		"v0": {
			version: V0,
			input: Entries{
				Entry{Key: []byte("key1"), Value: []byte("value1")},
				Entry{Key: []byte("key2"), Value: []byte("value2")},
				Entry{Key: []byte("key3"), Value: []byte("verylargevaluewithmorethan32byteslength")},
			},
			expected: common.Hash{
				0x71, 0x5, 0x2d, 0x48, 0x70, 0x46, 0x58, 0xa8, 0x43, 0x5f, 0xb9, 0xcb, 0xc7, 0xef, 0x69, 0xc7, 0x5d,
				0xad, 0x2f, 0x64, 0x0, 0x1c, 0xb3, 0xb, 0xfa, 0x1, 0xf, 0x7d, 0x60, 0x9e, 0x26, 0x57,
			},
		},
		"v1": {
			version: V1,
			input: Entries{
				Entry{Key: []byte("key1"), Value: []byte("value1")},
				Entry{Key: []byte("key2"), Value: []byte("value2")},
				Entry{Key: []byte("key3"), Value: []byte("verylargevaluewithmorethan32byteslength")},
			},
			expected: common.Hash{
				0x6a, 0x4a, 0x73, 0x27, 0x57, 0x26, 0x3b, 0xf2, 0xbc, 0x4e, 0x3, 0xa3, 0x41, 0xe3, 0xf8, 0xea, 0x63,
				0x5f, 0x78, 0x99, 0x6e, 0xc0, 0x6a, 0x6a, 0x96, 0x5d, 0x50, 0x97, 0xa2, 0x91, 0x1c, 0x29,
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			maxInline, err := testCase.version.Root(testCase.input)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, maxInline)
		})
	}
}
