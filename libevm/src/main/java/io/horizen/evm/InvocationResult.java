package io.horizen.evm;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.math.BigInteger;

public class InvocationResult {
    public final byte[] returnData;
    public final BigInteger leftOverGas;
    public final String executionError;

    public InvocationResult(
        @JsonProperty("returnData") byte[] returnData,
        @JsonProperty("leftOverGas") BigInteger leftOverGas,
        @JsonProperty("executionError") String executionError
    ) {
        this.returnData = returnData;
        this.leftOverGas = leftOverGas;
        this.executionError = executionError;
    }
}
