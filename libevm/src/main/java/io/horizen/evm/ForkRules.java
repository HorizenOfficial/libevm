package io.horizen.evm;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.math.BigInteger;

/*
This class contains which go-ethereum fork points should be activated.
The mapping between SDK fork points and go-ethereum fork points must be done inside the SDK.
Objects of this class must be passed to all EVM invocations that require checking a go-ethereum fork point.
 */
public class ForkRules {
    public final boolean isShanghai;

    public ForkRules(
        @JsonProperty("isShanghai") boolean isShanghai
    ) {
        this.isShanghai = isShanghai;
    }
}
