package io.horizen.evm;

import io.horizen.evm.params.DatabaseParams;

public abstract class Database extends ResourceHandle {
    public Database(int handle) {
        super(handle);
    }

    @Override
    public void close() {
        LibEvm.invoke("DatabaseClose", new DatabaseParams(handle));
    }
}
