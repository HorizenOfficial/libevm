package io.horizen.evm.params;

import java.math.BigInteger;

public class TracerTxStartParams extends TracerParams {
    public final BigInteger gasLimit;

    public TracerTxStartParams(int tracerHandle, BigInteger gasLimit) {
        super(tracerHandle);
        this.gasLimit = gasLimit;
    }
}
