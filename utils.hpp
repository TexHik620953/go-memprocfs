#ifdef __cplusplus
extern "C" {
#endif

#ifdef _WIN32
#include <stdint.h>
#endif //_WIN32

#include "vmmdll.h"
uint32_t FixCr3(VMM_HANDLE vmm_handle, DWORD target_pid, char* target_process_name);
#ifdef __cplusplus
}
#endif