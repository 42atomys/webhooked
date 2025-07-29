package integration_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SecurityIntegrationTestSuite struct {
	IntegrationTestSuite
}

func TestSecurityIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(SecurityIntegrationTestSuite))
}
