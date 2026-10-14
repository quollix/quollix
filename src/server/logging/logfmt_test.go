package logging

import (
	"strings"
	"testing"

	"github.com/quollix/common/assert"
)

func TestSplitLogfmtSuffix_Values(t *testing.T) {
	for _, tc := range []struct {
		line, message string
		fields        map[string]any
	}{
		{"plain text", "plain text", nil},
		{`empty= quoted_empty="" repeated=first repeated=last`, "", map[string]any{"empty": "", "quoted_empty": "", "repeated": "last"}},
		{" \t ", "", nil},
		{`ready a=b c="two words" a=last`, "ready", map[string]any{"a": "last", "c": "two words"}},
		{"a=\tquoted=\"\" _key.name-1=value", "", map[string]any{"a": "", "quoted": "", "_key.name-1": "value"}},
		{`msg="say \"hello\"" path="C:\\logs"`, "", map[string]any{"msg": `say "hello"`, "path": `C:\logs`}},
		{`msg="\b\f\n\r\t\/"`, "", map[string]any{"msg": "\b\f\n\r\t/"}},
		{`msg="caf\u00e9 \uD834\uDD1E"`, "", map[string]any{"msg": "café 𝄞"}},
		{`msg="\uD800x\uDC00\uD800\u0041"`, "", map[string]any{"msg": "�x��A"}},
		{`msg="café 𝄞"`, "", map[string]any{"msg": "café 𝄞"}},
		{`msg="a=b"`, "", map[string]any{"msg": "a=b"}},
		{`path=C:\logs`, "", map[string]any{"path": `C:\logs`}},
		{`1key=value ready a=b`, "1key=value ready", map[string]any{"a": "b"}},
		{`msg="\u0000"`, "", map[string]any{"msg": "\x00"}},
	} {
		t.Run(tc.line, func(t *testing.T) {
			message, fields, ok := splitLogfmtSuffix(tc.line)
			assert.True(t, ok)
			assert.Equal(t, tc.message, message)
			assert.Equal(t, tc.fields, fields)
		})
	}
}

func TestSplitLogfmtSuffix_MalformedSections(t *testing.T) {
	for _, line := range []string{
		`msg="\b\f\n\r\t\/\'"`,
		`ready a="unterminated`, `a="one"b=two`, `a=b bare`, `a=b =value`, `a=b 1key=value`,
		`a=b=c`, `a=bare"quote`, `a="\q"`, `a="\x41"`, `a="\U00000041"`, `a="\123"`,
		`a="\u"`, `a="\u123"`, `a="\uZZZZ"`, `a="\uD800\uZZZZ"`, `a="trailing\`,
		"a=one\ntwo", "a=one\rtwo", "a=one\x00two", "a=\"one\ntwo\"", "a=\"one\ttwo\"",
	} {
		t.Run(line, func(t *testing.T) {
			_, fields, ok := splitLogfmtSuffix(line)
			assert.False(t, ok)
			assert.Nil(t, fields)
		})
	}
}

func TestSplitLogfmtSuffix_LargeValue(t *testing.T) {
	value := strings.Repeat("x", 100)
	_, fields, ok := splitLogfmtSuffix(`msg="` + value + `"`)
	assert.True(t, ok)
	assert.Equal(t, value, fields["msg"])
}

func FuzzSplitLogfmtSuffix(f *testing.F) {
	for _, line := range []string{`a=b`, `prefix a="two words"`, `a="\uD834\uDD1E"`, `a="\uD800"`, `a="\u123`, `a="\q"`, `a="trailing\`} {
		f.Add(line)
	}
	f.Fuzz(func(t *testing.T, line string) {
		_, fields, ok := splitLogfmtSuffix(line)
		if !ok {
			assert.Nil(t, fields)
		}
		for key, value := range fields {
			assert.True(t, logfmtKey(key))
			_, isString := value.(string)
			assert.True(t, isString)
		}
	})
}
