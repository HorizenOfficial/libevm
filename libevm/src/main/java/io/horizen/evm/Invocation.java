package io.horizen.evm;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.math.BigInteger;

public class Invocation {
    public final Address caller;
    public final Address callee;
    public final BigInteger value;
    public final byte[] input;
    public final BigInteger gas;
    public final boolean readOnly;

    public Invocation(
        @JsonProperty("caller") Address caller,
        @JsonProperty("callee") Address callee,
        @JsonProperty("value") BigInteger value,
        @JsonProperty("input") byte[] input,
        @JsonProperty("gas") BigInteger gas,
        @JsonProperty("readOnly") boolean readOnly
    ) {
        this.caller = caller;
        this.callee = callee;
        this.value = value;
        this.input = input;
        this.gas = gas;
        this.readOnly = readOnly;
    }
}
