package users

import (
	"net/http/httptest"
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

func TestBuildSignInURLForRequest_PreservesCurrentRequestURIAsNext(t *testing.T) {
	request := httptest.NewRequest("GET", api.Paths.FrontendAppOpen+"?app=sampleapp&path=%2Fcustom-path%3Ftab%3Dfiles", nil)

	signInURL := BuildSignInURLForRequest(request)

	assert.Equal(t, api.Paths.FrontendSignIn+"?next=%2Fapp-open%3Fapp%3Dsampleapp%26path%3D%252Fcustom-path%253Ftab%253Dfiles", signInURL)
}
