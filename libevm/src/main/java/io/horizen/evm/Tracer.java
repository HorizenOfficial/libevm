package io.horizen.evm;

import io.horizen.evm.params.TracerCreateParams;
import io.horizen.evm.params.TracerParams;
import io.horizen.evm.results.TracerResult;

public class Tracer extends ResourceHandle {
    public Tracer(TraceOptions options) {
        super(LibEvm.invoke("TracerCreate", new TracerCreateParams(options), int.class));
    }

    @Override
    public void close() {
        LibEvm.invoke("TracerRemove", new TracerParams(handle));
    }

    public TracerResult getResult() {
        return LibEvm.invoke("TracerResult", new TracerParams(handle), TracerResult.class);
    }
}
