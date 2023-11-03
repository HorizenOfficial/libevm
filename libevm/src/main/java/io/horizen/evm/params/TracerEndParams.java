package io.horizen.evm.params;

import java.math.BigInteger;

public class TracerEndParams extends TracerParams {
    public final byte[] output;
    public final BigInteger gasUsed;
    public final long duration;
    public final String err;

    public TracerEndParams(int tracerHandle, byte[] output, BigInteger gasUsed, long duration, String err) {
        super(tracerHandle);
        this.output = output;
        this.gasUsed = gasUsed;
        this.duration = duration;
        this.err = err;
    }
}
