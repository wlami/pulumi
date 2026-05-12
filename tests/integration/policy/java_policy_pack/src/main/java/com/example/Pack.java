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
                .name("no-dynamic-resources")
                .description("Mandatory failing policy for Plan C MVP integration test.")
                .enforcementLevel(EnforcementLevel.MANDATORY)
                .validate((rArgs, report) -> {
                  if ("pulumi-nodejs:dynamic:Resource".equals(rArgs.type())) {
                    report.violation("dynamic Resource is not allowed (java integration test)");
                  }
                })
                .build())
            .build(),
        args);
  }
}
