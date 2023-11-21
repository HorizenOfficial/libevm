package io.horizen.evm.params;

import com.fasterxml.jackson.annotation.JsonUnwrapped;
import io.horizen.evm.TraceOptions;

public class TracerCreateParams {
    @JsonUnwrapped
    public final TraceOptions traceOptions;

    public TracerCreateParams(TraceOptions traceOptions) {
        this.traceOptions = traceOptions;
    }
}
