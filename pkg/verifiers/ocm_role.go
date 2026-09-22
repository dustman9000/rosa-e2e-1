package verifiers

import (
	"fmt"

	sdk "github.com/openshift-online/ocm-sdk-go"
)

// ocmRoleLabelKey is the organization label under which the linked OCM role ARN is stored.
// OCM roles are becoming mandatory for ROSA cluster operations (ROSA-637); a linked role is
// represented as an organization-scoped label carrying the IAM role ARN.
const ocmRoleLabelKey = "sts_ocm_role"

// VerifyOCMRoleLinked asserts that the caller's OCM organization has an OCM role linked, i.e.
// an organization label "sts_ocm_role" holding a non-empty IAM role ARN.
//
// OCM role linkage is mandatory for ROSA cluster operations (ROSA-637 / ROSAENG-64626): without
// it, cluster operations are rejected. This verifies the positive case for the org under test.
func VerifyOCMRoleLinked(conn *sdk.Connection) error {
	arn, err := GetLinkedOCMRoleARN(conn)
	if err != nil {
		return err
	}
	if arn == "" {
		return fmt.Errorf("OCM role is not linked to the organization (required by ROSA-637)")
	}
	return nil
}

// GetLinkedOCMRoleARN returns the ARN of the OCM role linked to the caller's organization, or an
// empty string if no OCM role is linked.
func GetLinkedOCMRoleARN(conn *sdk.Connection) (string, error) {
	acctResp, err := conn.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil {
		return "", fmt.Errorf("getting current account: %w", err)
	}

	org := acctResp.Body().Organization()
	if org == nil || org.ID() == "" {
		return "", fmt.Errorf("current account has no organization")
	}
	orgID := org.ID()

	labelsResp, err := conn.AccountsMgmt().V1().Organizations().Organization(orgID).Labels().List().Send()
	if err != nil {
		return "", fmt.Errorf("listing labels for organization %s: %w", orgID, err)
	}

	for _, label := range labelsResp.Items().Slice() {
		if label.Key() == ocmRoleLabelKey {
			return label.Value(), nil
		}
	}
	return "", nil
}
