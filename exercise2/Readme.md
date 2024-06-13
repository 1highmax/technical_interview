# Exercise 2

## Task Description
Shred tool in Go

Implement a Shred(path) function that will overwrite the given file (e.g. “randomfile”) 3 times with random data and delete the file afterwards. Note that the file may contain any type of data.  
You are expected to give information about the possible test cases for your Shred function, including the ones that  you don’t implement,  and implementing the full test coverage is a bonus :)  
In a few lines briefly discuss the possible use cases for such a helper function as well as advantages and drawbacks of addressing them with  this approach. 




## Assumptions
- Golang installed
- Internet connection available to download golang modules

## Usage 
Building:
```bash
go build
```
Running:
```bash
echo "test" > testfile
./shred-tool testfile
# File successfully shredded
```
Testing:
```bash
go test -coverpkg=shred-tool/shred -coverprofile=coverage.out ./... # -v for verbose
# Entropy test results for the random number generator are machine-specific
go tool cover -func=coverage.out
```
Coverage inspection in browser:
```bash
go tool cover -html=coverage.out
```

## Discussion
### Testing
- [shred/shred_test.go](shred/shred_test.go) tests the functionlity of [shred/shred.go](shred/shred.go), where it achieves 100% code coverage
- The most important test checks if a file is indeed overwritten 3 times, since this is important for security. For this purpose, the tool uses dependency injection for the file interface, and some test cases are executed with [shred/mockfile_test.go](shred/mockfile_test.go)
- Other implemented tests cover basic cases where files are not existing, unwriteable, deleted at the end, closed properly, empty or large (1M)
- [shred/entropy_test.go](shred/entropy_test.go) includes a basic chi-square entropy test for the quality of the random number generator. For highly sensible data, this should be performed not as a one-time test, but before every execution of the tool. I did not include such runtime functionality to not bloat up the exercise.
- The [main.go](main.go) file is not tested with respect to coverage, but executed through the `TestMainArgumentHandling` test case in [shred/shred_test.go](shred/shred_test.go) This maintains readability in the source code versus outsourcing even the last LOC to other functions.
- [shred/shred_test.go](shred/shred_test.go) skips "*Test 3: Permission Denied*" if the test is performed by root.  This is because some filesystem do not support `chattr -i`. Therefore, such filesystems have no trivial way to cause a permission error when root tries to overwrite a file. (i.e. Docker LayeredFS and APFS).

### Use Cases
#### Magnetic storage
- The tool can be used to securely erase data from magnetic storage, like HDDs, because for trivial overwrites, such storage mediums can retain the original data as small, residual magnetic forces at the physical sectors on the disk or tape. This is counteracted by repeatedly overwriting the sectors with random data.
- It can be necessary to securely erase magnetic data due to GDPR compliance, or simply as good measure to protect privacy or intellectual property
#### SSD
- For SSDs, this tool is useless, because wear-leveling algorithms distribute writes across many memory cells, so the original file may not get overwritten correctly. Additionally, SSDs have a limited lifespan in terms of writing access, making shredding algorithms a bad choice in general.
- There are other secure-erase tools for SATA and NVME SSDs