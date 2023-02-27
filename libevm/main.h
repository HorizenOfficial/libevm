
// callback function definition
typedef int (*callbackProxy)(int handle, char *args, char *buffer);

// used by GO to invoke the callback, as GO cannot invoke C function pointers
int invokeCallbackProxy(callbackProxy proxy, int handle, char *args, char* buffer);
