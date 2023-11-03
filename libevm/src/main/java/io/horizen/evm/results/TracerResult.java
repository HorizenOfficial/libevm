package io.horizen.evm.results;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.JsonNode;

public class TracerResult {
    public final JsonNode result;

    public TracerResult(@JsonProperty("result") JsonNode result) {
        this.result = result;
    }
}
