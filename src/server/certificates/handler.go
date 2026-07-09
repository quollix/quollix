package certificates

import (
	"net/http"
	"server/configs"
	"server/tools"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

var expectedWIldcardGenerationErrors = u.MapOf(CertificatChallengeNotPossibleForLocalhost)

type CertificateHandler struct {
	WildcardCertificateService WildcardCertificateService
	Config                     *tools.GlobalConfig
	ConfigsRepository          configs.ConfigsRepository
	CertificatePersister       CertificatePersister
	CertificateService         CertificateService
	OperationMonitor           OperationMonitor
}

func (s *CertificateHandler) WildcardCertificateGenerationHandler(w http.ResponseWriter, r *http.Request) {
	session, info, err := s.WildcardCertificateService.StartDns01Session()
	if err != nil {
		u.WriteResponseError(w, expectedWIldcardGenerationErrors, err)
		return
	}
	go s.WildcardCertificateService.FinishDns01Session(session, info.WildcardKeyAuth)

	u.SendJsonResponse(w, info)
}

func (s *CertificateHandler) CertificateUploadHandler(w http.ResponseWriter, r *http.Request) {
	certificateFile, ok := validation.ReadBody[tools.BinaryFile](w, r)
	if !ok {
		return
	}

	certBundle, err := NewCertificateBundleFromPemBytes(certificateFile.Content)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	if err := s.CertificateService.ReplaceCertificate(certBundle); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (s *CertificateHandler) CertificateDownloadHandler(w http.ResponseWriter, r *http.Request) {
	bundle, err := s.CertificateService.GetCurrentCertificate()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	u.SendJsonResponse(w, tools.BinaryFile{
		FileName: "certificate.pem",
		Content:  bundle.GetBytes(),
	})
}

func (s *CertificateHandler) ResetCertificateHandler(w http.ResponseWriter, r *http.Request) {
	certBundle, err := s.CertificateService.GenerateUniversalSelfSignedCert()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	if err := s.CertificateService.ReplaceCertificate(certBundle); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

type OperationStatus struct {
	State            string `json:"state"`
	CurrentOperation string `json:"current_operation"`
}

func (s *CertificateHandler) GetOperationMonitorStatus(w http.ResponseWriter, r *http.Request) {
	state, operation := s.OperationMonitor.GetStatus()

	u.SendJsonResponse(w, OperationStatus{
		State:            string(state),
		CurrentOperation: operation,
	})
}
