#ifdef __cplusplus
extern "C" {
#endif

#ifdef _WIN32
#include <stdint.h>
#endif //_WIN32

#include "vmmdll.h"
uint32_t FixCr3(VMM_HANDLE vmm_handle, DWORD target_pid, char* target_process_name);

// VAD map helpers (CGo cannot access flexible array members or bitfields)
static inline DWORD VadMap_GetCount(PVMMDLL_MAP_VAD pVadMap) {
    return pVadMap->cMap;
}

static inline PVMMDLL_MAP_VADENTRY VadMap_GetEntry(PVMMDLL_MAP_VAD pVadMap, DWORD i) {
    return &pVadMap->pMap[i];
}

static inline DWORD VadEntry_GetVadType(PVMMDLL_MAP_VADENTRY e) { return e->VadType; }
static inline DWORD VadEntry_GetProtection(PVMMDLL_MAP_VADENTRY e) { return e->Protection; }
static inline DWORD VadEntry_GetfImage(PVMMDLL_MAP_VADENTRY e) { return e->fImage; }
static inline DWORD VadEntry_GetfFile(PVMMDLL_MAP_VADENTRY e) { return e->fFile; }
static inline DWORD VadEntry_GetfPageFile(PVMMDLL_MAP_VADENTRY e) { return e->fPageFile; }
static inline DWORD VadEntry_GetfPrivateMemory(PVMMDLL_MAP_VADENTRY e) { return e->fPrivateMemory; }
static inline DWORD VadEntry_GetfTeb(PVMMDLL_MAP_VADENTRY e) { return e->fTeb; }
static inline DWORD VadEntry_GetfStack(PVMMDLL_MAP_VADENTRY e) { return e->fStack; }
static inline DWORD VadEntry_GetfHeap(PVMMDLL_MAP_VADENTRY e) { return e->fHeap; }
static inline DWORD VadEntry_GetHeapNum(PVMMDLL_MAP_VADENTRY e) { return e->HeapNum; }
static inline DWORD VadEntry_GetCommitCharge(PVMMDLL_MAP_VADENTRY e) { return e->CommitCharge; }
static inline DWORD VadEntry_GetMemCommit(PVMMDLL_MAP_VADENTRY e) { return e->MemCommit; }
static inline LPSTR VadEntry_GetText(PVMMDLL_MAP_VADENTRY e) { return e->uszText; }

#ifdef __cplusplus
}
#endif