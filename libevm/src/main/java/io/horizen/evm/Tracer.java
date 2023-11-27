package io.horizen.evm;

import io.horizen.evm.params.*;
import io.horizen.evm.results.TracerResult;

import java.math.BigInteger;

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

    // Transaction level
    public void CaptureTxStart(BigInteger gasLimit) {
        LibEvm.invoke("TracerCaptureTxStart", new TracerTxStartParams(handle, gasLimit));
    }

    // Transaction level
    public void CaptureTxEnd(BigInteger restGas) {
        LibEvm.invoke("TracerCaptureTxEnd", new TracerTxEndParams(handle, restGas));
    }

    // Top call frame
    public void CaptureStart(
        ResourceHandle stateDBHandle,
        EvmContext context,
        Address from,
        Address to,
        boolean create,
        byte[] input,
        BigInteger gas,
        BigInteger value
    ) {
        LibEvm.invoke(
            "TracerCaptureStart",
            new TracerStartParams(handle, stateDBHandle.handle, context, from, to, create, input, gas, value)
        );
    }

    // Top call frame
    public void CaptureEnd(byte[] output, BigInteger gasUsed, String err) {
        LibEvm.invoke("TracerCaptureEnd", new TracerEndParams(handle, output, gasUsed, err));
    }

    // Rest of call frames
    public void CaptureEnter(
        TracerOpCode opCode,
        Address from,
        Address to,
        byte[] input,
        BigInteger gas,
        BigInteger value
    ) {
        LibEvm.invoke(
            "TracerCaptureEnter",
            new TracerEnterParams(handle, opCode.getName(), from, to, input, gas, value)
        );
    }

    // Rest of call frames
    public void CaptureExit(byte[] output, BigInteger gasUsed, String err) {
        LibEvm.invoke("TracerCaptureExit", new TracerExitParams(handle, output, gasUsed, err));
    }
}
