package verifiers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	sdk "github.com/openshift-online/ocm-sdk-go"
	ocmerrors "github.com/openshift-online/ocm-sdk-go/errors"
)

// ocmRoleLabelKey is the organization label under which linked OCM role ARNs are stored.
//
// A linked OCM role is represented as an organization-scoped label whose value is a
// comma-separated list of IAM role ARNs (one entry per linked AWS account). OCM roles are
// mandatory for ROSA cluster operations (ROSA-637 / ROSAENG-64626).
//
// This mirrors the openshift/rosa CLI (pkg/ocm/helpers.go: OCMRoleLabel / GetOCMRoleARN /
// CheckIfAWSAccountExists), which fetches the label directly and treats HTTP 404 as
// "not linked".
const ocmRoleLabelKey = "sts_ocm_role"

// VerifyOCMRoleLinked asserts that the caller's OCM organization has at least one OCM role
// linked.
//
// Prefer VerifyOCMRoleLinkedForAccount when the test's AWS account ID is known: a non-empty
// label only proves that *some* account in the organization is linked, not the account under
// test. This weaker check is kept for cases where the AWS account ID is unavailable.
func VerifyOCMRoleLinked(conn *sdk.Connection) error {
	arns, err := GetLinkedOCMRoleARNs(conn)
	if err != nil {
		return err
	}
	if len(arns) == 0 {
		return fmt.Errorf("no OCM role is linked to the organization (required by ROSA-637)")
	}
	return nil
}

// VerifyOCMRoleLinkedForAccount asserts that an OCM role ARN belonging to awsAccountID is
// linked to the caller's organization.
//
// This matches rosa's CheckIfAWSAccountExists, which parses each ARN in the label and matches
// on the AWS account ID, because the org-scoped label aggregates ARNs across every linked
// account.
func VerifyOCMRoleLinkedForAccount(conn *sdk.Connection, awsAccountID string) error {
	if awsAccountID == "" {
		return fmt.Errorf("awsAccountID is required to verify OCM role linkage")
	}
	arns, err := GetLinkedOCMRoleARNs(conn)
	if err != nil {
		return err
	}
	for _, a := range arns {
		parsed, perr := arn.Parse(a)
		if perr != nil {
			continue
		}
		if parsed.AccountID == awsAccountID {
			return nil
		}
	}
	return fmt.Errorf(
		"no OCM role linked for AWS account %s (found %d linked role(s)); OCM role linkage is required by ROSA-637",
		awsAccountID, len(arns))
}

// GetLinkedOCMRoleARNs returns the IAM role ARNs of the OCM roles linked to the caller's
// organization, or an empty slice if none are linked.
//
// It fetches the organization "sts_ocm_role" label directly and treats HTTP 404 as "not
// linked" (matching rosa's behavior), rather than erroring.
func GetLinkedOCMRoleARNs(conn *sdk.Connection) ([]string, error) {
	acctResp, err := conn.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil {
		return nil, fmt.Errorf("getting current account: %w", err)
	}

	org := acctResp.Body().Organization()
	if org == nil || org.ID() == "" {
		return nil, fmt.Errorf("current account has no organization")
	}
	orgID := org.ID()

	labelResp, err := conn.AccountsMgmt().V1().
		Organizations().Organization(orgID).
		Labels().Labels(ocmRoleLabelKey).Get().Send()
	if err != nil {
		if ocmErr, ok := err.(*ocmerrors.Error); ok && ocmErr.Status() == http.StatusNotFound {
			// A 404 means the label is absent, i.e. no OCM role is linked.
			return nil, nil
		}
		return nil, fmt.Errorf("getting %s label for organization %s: %w", ocmRoleLabelKey, orgID, err)
	}

	value := labelResp.Body().Value()
	if value == "" {
		return nil, nil
	}

	var arns []string
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			arns = append(arns, trimmed)
		}
	}
	return arns, nil
}
