#include <stdlib.h>
#include "main.h"

// used by GO to invoke the callback, as GO cannot invoke C function pointers
char* invokeCallbackProxy(callbackProxy proxy, int handle, char *args) {
    if (proxy == NULL) return NULL;
    return proxy(handle, args);
}
