package io.horizen.evm.params;

import java.math.BigInteger;

public class TracerExitParams extends TracerParams {
    public final byte[] output;
    public final BigInteger gasUsed;
    public final String err;

    public TracerExitParams(int tracerHandle, byte[] output, BigInteger gasUsed, String err) {
        super(tracerHandle);
        this.output = output;
        this.gasUsed = gasUsed;
        this.err = err;
    }
}
