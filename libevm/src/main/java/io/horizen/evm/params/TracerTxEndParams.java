package io.horizen.evm.params;

import java.math.BigInteger;

public class TracerTxEndParams extends TracerParams {
    public final BigInteger restGas;

    public TracerTxEndParams(int tracerHandle, BigInteger restGas) {
        super(tracerHandle);
        this.restGas = restGas;
    }
}
