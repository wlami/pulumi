package com.example;

import com.pulumi.policy.EnforcementLevel;
import com.pulumi.policy.PolicyPack;
import com.pulumi.policy.PolicyPackArgs;
import com.pulumi.policy.ResourceValidationPolicy;

public class Pack {
  public static void main(String[] args) {
    PolicyPack.run("java-integration-test-pack",
        PolicyPackArgs.builder()
            .enforcementLevel(EnforcementLevel.MANDATORY)
            .policies(ResourceValidationPolicy.builder()
                .name("no-random-passwords")
                .description("Mandatory failing policy for Plan D Phase 2 integration test.")
                .enforcementLevel(EnforcementLevel.MANDATORY)
                .validate((rArgs, report) -> {
                  if ("random:index:RandomPassword".equals(rArgs.type())) {
                    report.violation("RandomPassword resources are not allowed (java integration test)");
                  }
                })
                .build())
            .build(),
        args);
  }
}
