package io.horizen.evm;

import com.fasterxml.jackson.annotation.JsonValue;

public abstract class ResourceHandle implements AutoCloseable {
    /**
     * Handle to a native resource that requires manual release.
     */
    @JsonValue
    final int handle;

    public ResourceHandle(int handle) {
        this.handle = handle;
    }

    @Override
    public String toString() {
        return String.format("ResourceHandle{handle=%d}", handle);
    }
}
