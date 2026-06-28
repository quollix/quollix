package certificates

const (
	SampleWildcardKeyAuth = "uJi0naVfLcobf2bK_t4-VkS0HFCK0U1WXkpGCGl3irE"
)

type WildcardCertificateServiceMock struct {
	OperationMonitor OperationMonitor
}

func (s *WildcardCertificateServiceMock) StartDns01Session() (*Dns01Session, *Dns01ChallengeInfo, error) {
	txtDnsRecordToCreate := &Dns01ChallengeInfo{
		RecordName:      acmeChallengePrefix + "localhost",
		WildcardKeyAuth: SampleWildcardKeyAuth,
	}
	return nil, txtDnsRecordToCreate, nil
}

func (s *WildcardCertificateServiceMock) FinishDns01Session(session *Dns01Session, wildcardKeyAuth string) {
	id := s.OperationMonitor.BeginRun("generating wildcard certificate for localhost")
	// do nothing
	s.OperationMonitor.EndRun(id, true, "finished sample wildcard certificate generation successfully")
}
