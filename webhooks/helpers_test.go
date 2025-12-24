package webhooks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildMetadata(t *testing.T) {
	meta := BuildMetadata("org_1", "conn_1", "pay_1")
	require.Equal(t, "org_1", meta[MetadataOrgID])
	require.Equal(t, "conn_1", meta[MetadataConnectionID])
	require.Equal(t, "pay_1", meta[MetadataPaymentID])

	empty := BuildMetadata("", "", "")
	require.Len(t, empty, 0)
}

func TestSignatures(t *testing.T) {
	body := []byte(`{"foo":"bar"}`)
	require.Equal(t, "", SignPaystack(body, ""))
	require.Equal(t, "f519630fca75a1fc5d62fd2547858feccd17a564fcc45c8e209b66cf1d7711b7ceec3e299961e6ae4b5c5f130dcacbe0fdb552bd2f9e9d9ecc06cc783681faa0", SignPaystack(body, "secret"))
	require.Equal(t, "3f3ab3986b656abb17af3eb1443ed6c08ef8fff9fea83915909d1b421aec89be", SignHubtel(body, "secret"))
	require.Equal(t, "3f3ab3986b656abb17af3eb1443ed6c08ef8fff9fea83915909d1b421aec89be", SignPolar(body, "secret"))
	require.Equal(t, "secret-hash", FlutterwaveHash(" secret-hash "))
}
