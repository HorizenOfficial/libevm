package io.horizen.evm;

import io.horizen.evm.params.LevelDBParams;

public class LevelDBDatabase extends Database {
    /**
     * Open a LevelDB instance in the given path.
     *
     * @param path data directory to pass to levelDB
     * @param preimages enable saving preimages. It should be used only when a state dump is requested.
     */
    public LevelDBDatabase(String path, boolean preimages) {
        super(LibEvm.invoke("DatabaseOpenLevelDB", new LevelDBParams(path, preimages), int.class));
    }

    public LevelDBDatabase(String path) {
        this(path, false);
    }

    @Override
    public String toString() {
        return String.format("LevelDBDatabase{handle=%d}", handle);
    }
}
