package io.horizen.evm.params;

public class DumpParams extends HandleParams {
    public final String dumpFile;

    public DumpParams(int handle, String dumpFile) {
        super(handle);
        this.dumpFile = dumpFile;
    }
}
