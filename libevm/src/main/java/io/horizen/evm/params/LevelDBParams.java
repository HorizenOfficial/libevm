package io.horizen.evm.params;

public class LevelDBParams {
    public final String path;
    public final boolean preimages;

    public LevelDBParams(String path, boolean preimages) {
        this.path = path;
        this.preimages = preimages;
    }
}
