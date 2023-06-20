package io.horizen.evm;

public enum TracerOpCode {
    CALL("CALL"),
    CALLCODE("CALLCODE"),
    DELEGATECALL("DELEGATECALL"),
    STATICCALL("STATICCALL"),
    CREATE("CREATE"),
    CREATE2("CREATE2");

    private final String name;

    TracerOpCode(String name) {
        this.name = name;
    }

    public String getName() {
        return name;
    }
}
