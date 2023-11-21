package io.horizen.evm;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.math.BigInteger;

public class ExternalInvocation extends Invocation {
    public final int depth;

    public ExternalInvocation(
        @JsonProperty("caller") Address caller,
        @JsonProperty("callee") Address callee,
        @JsonProperty("value") BigInteger value,
        @JsonProperty("input") byte[] input,
        @JsonProperty("gas") BigInteger gas,
        @JsonProperty("readOnly") boolean readOnly,
        @JsonProperty("depth") int depth
    ) {
        super(caller, callee, value, input, gas, readOnly);
        this.depth = depth;
    }
}
