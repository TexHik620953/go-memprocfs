## Example initialization
```go
func main() {
	mfs, err := memprocfs.New("-device", "fpga", "-memmap", "memmap.txt")
	if err != nil {
		log.Fatalf("failed to create mfs %s", err.Error())
	}
	defer mfs.Close()

	pid, err = mfs.GetPidByName("RustClient.exe")
	if err != nil {
		log.Fatalf("failed to find pid %s", err.Error())
	}
	base, err = mfs.GetModuleBase(pid, "GameAssembly.dll")
	if err != nil {
		log.Fatalf("failed to get module base %s", err.Error())
	}
    fmt.Println("Base addr is: ", base)
}
```


## Example reading memory
```go
// Chain pointer read
addr, err := mfs.ReadPtr(pid, []uintptr{base + 0x123, 0x321, 0xAAAA, 0xBBBB, ...})

// Chain int read
value, err := mfs.ReadInt16(pid, []uintptr{base + 0x123, 0x321, 0xAAAA, 0xBBBB, ...})
value, err := mfs.ReadInt32(pid, []uintptr{base + 0x123, 0x321, 0xAAAA, 0xBBBB, ...})
value, err := mfs.ReadInt64(pid, []uintptr{base + 0x123, 0x321, 0xAAAA, 0xBBBB, ...})

// Chain uint read
value, err := mfs.ReadUInt32(pid, []uintptr{base + 0x123, 0x321, 0xAAAA, 0xBBBB, ...})
value, err := mfs.ReadUInt32(pid, []uintptr{base + 0x123, 0x321, 0xAAAA, 0xBBBB, ...})


// Chain read floats
value, err := mfs.ReadFloat32(pid, []uintptr{base + 0x123, 0x321, 0xAAAA, 0xBBBB, ...})
value, err := mfs.ReadFloat64(pid, []uintptr{base + 0x123, 0x321, 0xAAAA, 0xBBBB, ...})

// Chain read bool
value, err := mfs.ReadBools(pid, []uintptr{base + 0x123, 0x321, 0xAAAA, 0xBBBB, ...})

// Read continious memory
buffer, err := mfs.MemRead(pid, uintptr(0xABCDEF), int32(bytes2read))


```


## Example scatter reading
```go
// Create scatter
scatter, err := mfs.NewScatterTask(pid)
if err != nil {
	return nil, err
}

//Create some array for results
aas := make([]uintptr, actorsCount, actorsCount)

// Add read tasks
for i := 0; i < int(actorsCount); i++ {
scatter.AddRead(aaArr+uintptr(i)*0x8, 8, unsafe.Pointer(&aas[i]))
}

// Execute tasks
err = scatter.Execute()
if err != nil {
	return nil, err
}
// After executing read, underlying scatter being reset, so you can  use scatter again
```


## Example build task
```json
{
            "label": "go: build (linux)",
            "type": "shell",
            "command": "go",
            "args": [
                "build",
				"-o",
                "${workspaceRoot}/__debug_bin",
				"${workspaceRoot}/main.go"
            ],
            "options": {
				"env": {
					"CGO_ENABLED":"1",
					"GOOS":"linux",
					"GOARCH":"amd64",
					"CGO_CFLAGS": "-O0 -D _UNICODE -D UNICODE -D LINUX -I ./includes",
                    "CGO_CXXFLAGS": "-O0 -D _UNICODE -D UNICODE -D LINUX -I ./includes",
					"CGO_LDFLAGS": "-L ./libs/leechcore.so  ./libs/vmm.so"
				},
                "cwd": "${workspaceRoot}"
            },
			"dependsOn": [
				"clear (linux)"
			]
        }

```

