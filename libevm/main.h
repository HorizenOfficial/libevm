
// callback function definition
typedef char* (*callbackProxy)(int handle, char *args);

// used by GO to invoke the callback, as GO cannot invoke C function pointers
char* invokeCallbackProxy(callbackProxy proxy, int handle, char *args);
