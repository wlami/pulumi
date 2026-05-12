//go:build java_policy_integration

// Run with:
//
//	cd tests && PATH=$HOME/.pulumi/bin:$PATH go test -tags java_policy_integration \
//	  ./integration -run TestJavaPolicyPack -v -timeout 10m
//
// Requires: pulumi-language-java (Plan B fork) reachable via PATH before any
// system-installed pulumi-language-java; com.pulumi:pulumi-policy:0.1.0-SNAPSHOT
// in ~/.m2 (Plan D Phase 1 build); mvn on PATH, JDK 11+.

// Copyright 2026, Pulumi Corporation.  All rights reserved.

package ints

import (
	"os/exec"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	ptesting "github.com/pulumi/pulumi/sdk/v3/go/common/testing"
	"github.com/stretchr/testify/require"
)

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

	e.ImportDirectory("java_policy_program")
	e.ImportDirectory("policy")

	e.RunCommand("pulumi", "login", "--cloud-url", e.LocalURL())

	stackName, err := resource.NewUniqueHex("java-policy-test-", 8, -1)
	contract.AssertNoErrorf(err, "resource.NewUniqueHex should not fail with no maximum length set")
	e.RunCommand("pulumi", "stack", "init", stackName)

	stdout, stderr, err := e.GetCommandResults(
		"pulumi", "preview",
		"--policy-pack", "java_policy_pack",
	)
	t.Logf("preview stdout:\n%s", stdout)
	t.Logf("preview stderr:\n%s", stderr)

	require.Error(t, err, "preview should fail because the mandatory policy fires")
	combined := stdout + stderr
	require.Contains(t, combined, "no-random-passwords",
		"expected the policy name in the output, got: %s", combined)
	require.Contains(t, combined, "RandomPassword resources are not allowed",
		"expected the violation message in the output, got: %s", combined)
}
