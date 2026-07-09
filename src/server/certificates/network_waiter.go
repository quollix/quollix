package certificates

import (
	"context"
	"fmt"
	"net"
	"server/tools"
	"strings"
	"time"

	u "github.com/quollix/common/utils"
	"golang.org/x/crypto/acme"
)

const acmeChallengePrefix = "_acme-challenge."

type NetworkWaiter interface {
	Sleep(duration time.Duration)
	Retry(operationName string, frequency time.Duration, maxTime time.Duration, functionToRetry func() error) error
	LookupTxt(recordName string) ([]string, error)
	WaitForDnsTxtRecord(session *Dns01Session, wantedDnsTextRecord string) error
	WaitForAuthorization(wildcardAuthzUrl string) error
}

type NetworkWaiterImpl struct {
	AcmeClient       AcmeClient
	OperationMonitor OperationMonitor
}

func (n *NetworkWaiterImpl) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

func (n *NetworkWaiterImpl) Retry(operationName string, frequency time.Duration, maxTime time.Duration, functionToRetry func() error) error {
	deadline := time.Now().Add(maxTime)
	for {
		u.Logger.Info("Retrying operation", tools.OperationField, operationName)
		if err := functionToRetry(); err == nil {
			u.Logger.Info("Retry successful of operation", tools.OperationField, operationName)
			return nil
		} else if time.Now().After(deadline) {
			return u.Logger.NewError("retry deadline exceeded for operation", tools.OperationField, operationName)
		} else {
			u.Logger.Info("Attempt failed for operation, waiting...", tools.OperationField, operationName, "error_details", u.ExtractError(err))
			n.Sleep(frequency)
		}
	}
}

func (n *NetworkWaiterImpl) LookupTxt(recordName string) ([]string, error) {
	ctx := context.Background()
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network string, address string) (net.Conn, error) {
			dialer := net.Dialer{}
			return dialer.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}
	return resolver.LookupTXT(ctx, recordName)
}

func (n *NetworkWaiterImpl) WaitForDnsTxtRecord(session *Dns01Session, wantedDnsTextRecord string) error {
	var attempt = 0
	recordName := acmeChallengePrefix + session.Host
	return n.Retry("Poll DNS for challenge TXT records", 1*time.Minute, 20*time.Minute, func() error {

		attempt++
		operation := fmt.Sprintf("waiting for you to create the required DNS record, attempt %d", attempt)
		n.OperationMonitor.SetOperation(session.Id, operation)

		txtRecords, err := n.LookupTxt(recordName)
		if err != nil {
			return err
		}
		for _, txtRecord := range txtRecords {
			if strings.Contains(txtRecord, wantedDnsTextRecord) {
				return nil
			}
		}
		return u.Logger.NewError("DNS TXT record was found but value was not correct")
	})
}

func (n *NetworkWaiterImpl) WaitForAuthorization(wildcardAuthzUrl string) error {
	return n.Retry("Poll Let's Encrypt for DNS challenge validation", 1*time.Minute, 5*time.Minute, func() error {
		wildcardAuthorization, err := n.AcmeClient.GetAuthorization(wildcardAuthzUrl)
		if err != nil {
			return u.Logger.NewError(err.Error())
		}
		if wildcardAuthorization.Status == acme.StatusInvalid {
			return u.Logger.NewError("challenge validation failed")
		}
		if wildcardAuthorization.Status == acme.StatusValid {
			return nil
		}
		return u.Logger.NewError("challenge validation not yet complete")
	})
}
