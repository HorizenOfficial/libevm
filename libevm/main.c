#include <stdlib.h>
#include "main.h"

// used by GO to invoke the callback, as GO cannot invoke C function pointers
int invokeCallbackProxy(callbackProxy proxy, int handle, char *args, char *buffer) {
    if (proxy == NULL) return 0;
    return proxy(handle, args, buffer);
}
