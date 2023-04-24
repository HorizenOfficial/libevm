package io.horizen.evm;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.math.BigInteger;
import java.util.Objects;

public class InvocationResult {
    public final byte[] returnData;
    public final BigInteger leftOverGas;
    public final String executionError;
    public final boolean reverted;
    public final Address contractAddress;

    public InvocationResult(
        @JsonProperty("returnData") byte[] returnData,
        @JsonProperty("leftOverGas") BigInteger leftOverGas,
        @JsonProperty("executionError") String executionError,
        @JsonProperty("reverted") boolean reverted,
        @JsonProperty("contractAddress") Address contractAddress
    ) {
        this.returnData = Objects.requireNonNullElse(returnData, new byte[0]);
        this.leftOverGas = leftOverGas;
        this.executionError = executionError;
        this.reverted = reverted;
        this.contractAddress = contractAddress;
    }
}
