#include <thread>
#include <chrono>
#include <vector>
#include <iostream>
#include <sstream>
#include "utils.hpp"
#include <memory>
#include <cstring>

#define LOG(fmt, ...) std::printf(fmt, ##__VA_ARGS__)
struct Info
{
    uint32_t index;
    uint32_t process_id;
    uint64_t dtb;
    uint64_t kernelAddr;
    std::string name;
};

uint64_t cbSize = 0x80000;

VOID cbAddFile(_Inout_ HANDLE h, _In_ LPCSTR uszName, _In_ ULONG64 cb, _In_opt_ PVMMDLL_VFS_FILELIST_EXINFO pExInfo)
{
	if (strcmp(uszName, "dtb.txt") == 0)
		cbSize = cb;
}

uint32_t FixCr3(VMM_HANDLE vmm_handle, DWORD target_pid, char* target_process_name)
{
	PVMMDLL_MAP_MODULEENTRY module_entry = NULL;
	bool result = VMMDLL_Map_GetModuleFromNameU(vmm_handle, target_pid, const_cast<LPSTR>(target_process_name), &module_entry, 0);
	if (result)
		return 1; //Doesn't need to be patched lol

	if (!VMMDLL_InitializePlugins(vmm_handle))
	{
		LOG("[-] Failed VMMDLL_InitializePlugins call\n");
		return 0;
	}

	//have to sleep a little or we try reading the file before the plugin initializes fully
	std::this_thread::sleep_for(std::chrono::milliseconds(500));

	int retries = 0;

	for(retries = 0;retries<11;retries++)
	{
		BYTE bytes[4] = {0};
		DWORD i = 0;
		auto nt = VMMDLL_VfsReadU(vmm_handle, reinterpret_cast<LPCSTR>("\\misc\\procinfo\\progress_percent.txt"), bytes, 3, &i, 0);
		if (nt == VMMDLL_STATUS_SUCCESS && atoi(reinterpret_cast<LPSTR>(bytes)) == 100)
			break;

		std::this_thread::sleep_for(std::chrono::milliseconds(100));
	}
	if (retries >=9) {
		LOG("[-] Failed VMMDLL_VfsReadU after 9 retries\n");
		return 0;
	}

	VMMDLL_VFS_FILELIST2 VfsFileList;
	VfsFileList.dwVersion = VMMDLL_VFS_FILELIST_VERSION;
	VfsFileList.h = 0;
	VfsFileList.pfnAddDirectory = 0;
	VfsFileList.pfnAddFile = cbAddFile; //dumb af callback who made this system

	result = VMMDLL_VfsListU(vmm_handle, reinterpret_cast<LPCSTR>("\\misc\\procinfo\\"), &VfsFileList);
	if (!result)
		return 0;

	//read the data from the txt and parse it
	const size_t buffer_size = cbSize;
	std::unique_ptr<BYTE[]> bytes(new BYTE[buffer_size]);
	DWORD j = 0;
	auto nt = VMMDLL_VfsReadU(vmm_handle, reinterpret_cast<LPCSTR>("\\misc\\procinfo\\dtb.txt"), bytes.get(), buffer_size - 1, &j, 0);
	if (nt != VMMDLL_STATUS_SUCCESS)
		return 0;

	std::vector<uint64_t> possible_dtbs = { };
	std::string lines(reinterpret_cast<char*>(bytes.get()));
	std::istringstream iss(lines);
	std::string line = "";

	while (std::getline(iss, line))
	{
		Info info = { };

		std::istringstream info_ss(line);
		if (info_ss >> std::hex >> info.index >> std::dec >> info.process_id >> std::hex >> info.dtb >> info.kernelAddr >> info.name)
		{
			if (info.process_id == 0) //parts that lack a name or have a NULL pid are suspects
				possible_dtbs.push_back(info.dtb);
			if (std::string(target_process_name).find(info.name) != std::string::npos)
				possible_dtbs.push_back(info.dtb);
		}
	}

	//loop over possible dtbs and set the config to use it til we find the correct one
	for (size_t i = 0; i < possible_dtbs.size(); i++)
	{
		auto dtb = possible_dtbs[i];
		VMMDLL_ConfigSet(vmm_handle, VMMDLL_OPT_PROCESS_DTB | target_pid, dtb);
		result = VMMDLL_Map_GetModuleFromNameU(vmm_handle, target_pid, const_cast<LPCSTR>(target_process_name), &module_entry, 0);
		if (result)
		{
			LOG("[+] Patched DTB\n");
			return 1;
		}
	}

	LOG("[-] Failed to patch module\n");
	return 0;
}
