//go:build java_policy_integration

// Run with:
//
//	go test -tags java_policy_integration ./tests/integration -run TestJavaPolicyPack -v -timeout 10m
//
// Requires:
//   - pulumi CLI built from the Plan C fork (this repo) and on PATH
//   - pulumi-language-java built from the Plan B fork and on PATH
//   - com.pulumi:pulumi-policy:0.1.0-SNAPSHOT installed in ~/.m2
//   - mvn on PATH, JDK 11+
//   - node + yarn for the single_resource Pulumi program
//
// Copyright 2026, Pulumi Corporation.  All rights reserved.

package ints

import (
	"os/exec"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	ptesting "github.com/pulumi/pulumi/sdk/v3/go/common/testing"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/stretchr/testify/require"
)

// TestJavaPolicyPack runs `pulumi preview --policy-pack java_policy_pack`
// against the shared single_resource fixture and asserts the mandatory Java
// policy fires.
func TestJavaPolicyPack(t *testing.T) {
	if _, err := exec.LookPath("mvn"); err != nil {
		t.Skip("mvn not on PATH; skipping Java policy pack integration test")
	}
	if _, err := exec.LookPath("pulumi-language-java"); err != nil {
		t.Skip("pulumi-language-java not on PATH; build it from the Plan B fork first")
	}

	t.Parallel()

	e := ptesting.NewEnvironment(t)
	defer e.DeleteIfNotFailed()

	e.ImportDirectory("single_resource")
	e.ImportDirectory("policy")

	e.RunCommand("pulumi", "login", "--cloud-url", e.LocalURL())

	stackName, err := resource.NewUniqueHex("java-policy-test-", 8, -1)
	contract.AssertNoErrorf(err, "resource.NewUniqueHex should not fail with no maximum length set")
	e.RunCommand("pulumi", "stack", "init", stackName)

	// Wire up the single_resource Pulumi program (TypeScript).
	e.RunCommandWithRetry("yarn", "link", "@pulumi/pulumi")
	e.RunCommandWithRetry("yarn", "install")

	// The policy pack is built on demand by pulumi-language-java's policy mode
	// (mvn compile exec:java). No explicit build step needed here.

	// Run preview with the Java policy pack. Expect a non-zero exit because
	// the policy is MANDATORY and the single_resource program registers a
	// pulumi-nodejs:dynamic:Resource which violates the policy.
	stdout, stderr, err := e.GetCommandResults(
		"pulumi", "preview",
		"--policy-pack", "java_policy_pack",
	)
	t.Logf("preview stdout:\n%s", stdout)
	t.Logf("preview stderr:\n%s", stderr)

	require.Error(t, err, "preview should fail because the mandatory policy fires")
	combined := stdout + stderr
	require.Contains(t, combined, "no-dynamic-resources",
		"expected the policy name in the output, got: %s", combined)
	require.Contains(t, combined, "dynamic Resource is not allowed",
		"expected the violation message in the output, got: %s", combined)
}
