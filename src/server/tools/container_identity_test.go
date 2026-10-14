package tools

import (
	"testing"

	"github.com/quollix/common/assert"
)

func TestContainerIdentity_ParsesMatchingNames(t *testing.T) {
	for _, test := range []struct {
		name  string
		valid bool
	}{
		{"abc_def_ghi", true},
		{"/abc_def_ghi", true},
		{"abcdefghijklmnopqrst_def_ghi", true},
		{"ab_def_ghi", false},
		{"abcdefghijklmnopqrstu_def_ghi", false},
		{"abc_def_ghi_extra", false},
		{"abc_def_gh1", false},
		{"abc_Def_ghi", false},
		{"//abc_def_ghi", false},
		{"", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			identity, valid := ParseContainerName(test.name)
			assert.Equal(t, test.valid, valid)
			if !valid {
				assert.Nil(t, identity)
			}
		})
	}
	identity, valid := ParseContainerName("/quollix_sampleapp_web")
	assert.True(t, valid)
	assert.Equal(t, &ContainerIdentity{Maintainer: "quollix", App: "sampleapp", Service: "web"}, identity)
}
