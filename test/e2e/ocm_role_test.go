//go:build E2Etests

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openshift-online/rosa-e2e/pkg/labels"
	"github.com/openshift-online/rosa-e2e/pkg/verifiers"
)

// OCM roles are mandatory for ROSA cluster operations (ROSA-637). These specs validate that the
// organization under test has an OCM role linked, so cluster operations succeed after enforcement.
// See ROSAENG-64626.
var _ = Describe("Management Plane: OCM Role Linkage", labels.High, labels.Positive, labels.HCP, labels.Classic, labels.ManagementPlane, func() {
	It("should have an OCM role linked to the organization", func(ctx context.Context) {
		By("Verifying the organization has an OCM role linked (sts_ocm_role)")
		Expect(verifiers.VerifyOCMRoleLinked(conn)).To(Succeed())
	})

	It("should expose the linked OCM role ARN", func(ctx context.Context) {
		By("Resolving the linked OCM role ARN")
		arn, err := verifiers.GetLinkedOCMRoleARN(conn)
		Expect(err).NotTo(HaveOccurred())
		Expect(arn).To(HavePrefix("arn:aws:iam::"), "linked OCM role should be a valid IAM role ARN")
	})
})
